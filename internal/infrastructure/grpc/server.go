package grpc

import (
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/nurashi/Shipment-gRPC-microservice/gen/shipment"
)

type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

func NewServer(handler *Handler, addr string) (*Server, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterShipmentServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		listener:   lis,
	}, nil
}

func (s *Server) Start() error {
	log.Printf("gRPC server listening on %s", s.listener.Addr().String())
	return s.grpcServer.Serve(s.listener)
}

func (s *Server) Stop() {
	log.Println("shutting down gRPC server...")
	s.grpcServer.GracefulStop()
}
