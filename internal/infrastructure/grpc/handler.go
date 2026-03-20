package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/nurashi/Shipment-gRPC-microservice/gen/shipment"
	"github.com/nurashi/Shipment-gRPC-microservice/internal/application"
	"github.com/nurashi/Shipment-gRPC-microservice/internal/domain"
)

type Handler struct {
	pb.UnimplementedShipmentServiceServer
	service application.ShipmentService
}

func NewHandler(service application.ShipmentService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateShipment(ctx context.Context, req *pb.CreateShipmentRequest) (*pb.Shipment, error) {
	if req.GetReferenceNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "reference_number is required")
	}
	if req.GetOrigin() == "" {
		return nil, status.Error(codes.InvalidArgument, "origin is required")
	}
	if req.GetDestination() == "" {
		return nil, status.Error(codes.InvalidArgument, "destination is required")
	}
	if req.GetDriverName() == "" {
		return nil, status.Error(codes.InvalidArgument, "driver_name is required")
	}
	if req.GetUnitNumber() == "" {
		return nil, status.Error(codes.InvalidArgument, "unit_number is required")
	}
	if req.GetAmount() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than zero")
	}
	if req.GetDriverRevenue() < 0 {
		return nil, status.Error(codes.InvalidArgument, "driver_revenue must be non-negative")
	}

	input := domain.CreateShipmentInput{
		ReferenceNumber: req.GetReferenceNumber(),
		Origin:          req.GetOrigin(),
		Destination:     req.GetDestination(),
		DriverName:      req.GetDriverName(),
		UnitNumber:      req.GetUnitNumber(),
		Amount:          req.GetAmount(),
		DriverRevenue:   req.GetDriverRevenue(),
	}

	shipment, err := h.service.CreateShipment(ctx, input)
	if err != nil {
		return nil, mapDomainError(err)
	}

	return shipmentToProto(shipment), nil
}

func (h *Handler) GetShipment(ctx context.Context, req *pb.GetShipmentRequest) (*pb.Shipment, error) {
	if req.GetId() == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	shipment, err := h.service.GetShipment(ctx, req.GetId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return shipmentToProto(shipment), nil
}

func (h *Handler) AddShipmentEvent(ctx context.Context, req *pb.AddShipmentEventRequest) (*pb.ShipmentEvent, error) {
	if req.GetShipmentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "shipment_id is required")
	}
	if req.GetStatus() == pb.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED {
		return nil, status.Error(codes.InvalidArgument, "status is required")
	}

	domainStatus := protoStatusToDomain(req.GetStatus())

	event, err := h.service.AddShipmentEvent(ctx, req.GetShipmentId(), domainStatus, req.GetNotes())
	if err != nil {
		return nil, mapDomainError(err)
	}

	return eventToProto(event), nil
}

func (h *Handler) GetShipmentHistory(ctx context.Context, req *pb.GetShipmentHistoryRequest) (*pb.ShipmentHistoryResponse, error) {
	if req.GetShipmentId() == "" {
		return nil, status.Error(codes.InvalidArgument, "shipment_id is required")
	}

	events, err := h.service.GetShipmentHistory(ctx, req.GetShipmentId())
	if err != nil {
		return nil, mapDomainError(err)
	}

	pbEvents := make([]*pb.ShipmentEvent, len(events))
	for i, e := range events {
		pbEvents[i] = eventToProto(e)
	}

	return &pb.ShipmentHistoryResponse{Events: pbEvents}, nil
}

func shipmentToProto(s *domain.Shipment) *pb.Shipment {
	return &pb.Shipment{
		Id:              s.ID,
		ReferenceNumber: s.ReferenceNumber,
		Origin:          s.Origin,
		Destination:     s.Destination,
		Status:          domainStatusToProto(s.CurrentStatus),
		DriverName:      s.DriverName,
		UnitNumber:      s.UnitNumber,
		Amount:          s.Amount,
		DriverRevenue:   s.DriverRevenue,
		CreatedAt:       timestamppb.New(s.CreatedAt),
		UpdatedAt:       timestamppb.New(s.UpdatedAt),
	}
}

func eventToProto(e *domain.ShipmentEvent) *pb.ShipmentEvent {
	return &pb.ShipmentEvent{
		Id:         e.ID,
		ShipmentId: e.ShipmentID,
		Status:     domainStatusToProto(e.Status),
		Notes:      e.Notes,
		OccurredAt: timestamppb.New(e.OccurredAt),
	}
}

var domainToProtoStatus = map[domain.Status]pb.ShipmentStatus{
	domain.StatusPending:   pb.ShipmentStatus_SHIPMENT_STATUS_PENDING,
	domain.StatusPickedUp:  pb.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP,
	domain.StatusInTransit: pb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT,
	domain.StatusDelivered: pb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED,
	domain.StatusCancelled: pb.ShipmentStatus_SHIPMENT_STATUS_CANCELLED,
}

var protoToDomainStatus = map[pb.ShipmentStatus]domain.Status{
	pb.ShipmentStatus_SHIPMENT_STATUS_PENDING:    domain.StatusPending,
	pb.ShipmentStatus_SHIPMENT_STATUS_PICKED_UP:  domain.StatusPickedUp,
	pb.ShipmentStatus_SHIPMENT_STATUS_IN_TRANSIT: domain.StatusInTransit,
	pb.ShipmentStatus_SHIPMENT_STATUS_DELIVERED:  domain.StatusDelivered,
	pb.ShipmentStatus_SHIPMENT_STATUS_CANCELLED:  domain.StatusCancelled,
}

func domainStatusToProto(s domain.Status) pb.ShipmentStatus {
	if v, ok := domainToProtoStatus[s]; ok {
		return v
	}
	return pb.ShipmentStatus_SHIPMENT_STATUS_UNSPECIFIED
}

func protoStatusToDomain(s pb.ShipmentStatus) domain.Status {
	if v, ok := protoToDomainStatus[s]; ok {
		return v
	}
	return ""
}

func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrShipmentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidStatusTransition):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrShipmentTerminated):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrDuplicateReferenceNumber):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrMissingRequiredField):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrInvalidFieldValue):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
