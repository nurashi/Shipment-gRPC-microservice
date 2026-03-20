package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nurashi/Shipment-gRPC-microservice/config"
	"github.com/nurashi/Shipment-gRPC-microservice/internal/application"
	grpcserver "github.com/nurashi/Shipment-gRPC-microservice/internal/infrastructure/grpc"
	"github.com/nurashi/Shipment-gRPC-microservice/internal/infrastructure/persistence/postgres"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.GetPostgresConnString())
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	log.Println("connected to database")

	if err := postgres.RunMigrations(cfg.GetPostgresConnString(), "migrations"); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	shipmentRepo := postgres.NewShipmentRepo(pool)
	eventRepo := postgres.NewEventRepo(pool)

	svc := application.NewShipmentService(shipmentRepo, eventRepo)

	handler := grpcserver.NewHandler(svc)

	server, err := grpcserver.NewServer(handler, ":"+cfg.GRPCPort)
	if err != nil {
		log.Fatalf("failed to create gRPC server: %v", err)
	}

	// Graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		server.Stop()
		cancel()
	}()

	if err := server.Start(); err != nil {
		log.Fatalf("gRPC server error: %v", err)
	}
}
