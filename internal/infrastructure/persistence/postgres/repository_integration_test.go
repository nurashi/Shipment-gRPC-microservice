//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/nurashi/Shipment-gRPC-microservice/internal/domain"
	"github.com/nurashi/Shipment-gRPC-microservice/internal/infrastructure/persistence/postgres"
)

// --- Setup helpers ---

func testConnString() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		envOrDefault("DB_USER", "shipment"),
		envOrDefault("DB_PASSWORD", "shipment"),
		envOrDefault("DB_HOST", "localhost"),
		envOrDefault("DB_PORT", "5432"),
		envOrDefault("DB_NAME", "shipment_db"),
		envOrDefault("DB_SSLMODE", "disable"),
	)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func migrationsDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "../../../../migrations")
}

func setupDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, testConnString())
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Fatalf("failed to ping test database: %v", err)
	}

	if err := postgres.RunMigrations(testConnString(), migrationsDir()); err != nil {
		pool.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	t.Cleanup(func() {
		pool.Exec(ctx, "DELETE FROM shipment_events")
		pool.Exec(ctx, "DELETE FROM shipments")
		pool.Close()
	})

	return pool
}

func testInput(ref string) domain.CreateShipmentInput {
	return domain.CreateShipmentInput{
		ReferenceNumber: ref,
		Origin:          "Almaty",
		Destination:     "Astana",
		DriverName:      "Test Driver",
		UnitNumber:      "TRUCK-01",
		Amount:          1000.00,
		DriverRevenue:   300.00,
	}
}

// --- ShipmentRepo tests ---

func TestIntegration_ShipmentRepo_SaveAndFindByID(t *testing.T) {
	pool := setupDB(t)
	repo := postgres.NewShipmentRepo(pool)
	ctx := context.Background()

	shipment, err := domain.NewShipment(testInput("INT-001"))
	if err != nil {
		t.Fatalf("failed to build shipment: %v", err)
	}

	if err := repo.Save(ctx, shipment); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	found, err := repo.FindByID(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("FindByID failed: %v", err)
	}

	if found.ID != shipment.ID {
		t.Errorf("ID: expected %s, got %s", shipment.ID, found.ID)
	}
	if found.ReferenceNumber != "INT-001" {
		t.Errorf("ReferenceNumber: expected INT-001, got %s", found.ReferenceNumber)
	}
	if found.Origin != "Almaty" {
		t.Errorf("Origin: expected Almaty, got %s", found.Origin)
	}
	if found.Destination != "Astana" {
		t.Errorf("Destination: expected Astana, got %s", found.Destination)
	}
	if found.CurrentStatus != domain.StatusPending {
		t.Errorf("CurrentStatus: expected PENDING, got %s", found.CurrentStatus)
	}
	if found.Amount != 1000.00 {
		t.Errorf("Amount: expected 1000.00, got %f", found.Amount)
	}
}

func TestIntegration_ShipmentRepo_FindByID_NotFound(t *testing.T) {
	pool := setupDB(t)
	repo := postgres.NewShipmentRepo(pool)
	ctx := context.Background()

	_, err := repo.FindByID(ctx, "00000000-0000-0000-0000-000000000000")
	if !errors.Is(err, domain.ErrShipmentNotFound) {
		t.Errorf("expected ErrShipmentNotFound, got %v", err)
	}
}

func TestIntegration_ShipmentRepo_ExistsByReferenceNumber(t *testing.T) {
	pool := setupDB(t)
	repo := postgres.NewShipmentRepo(pool)
	ctx := context.Background()

	shipment, _ := domain.NewShipment(testInput("INT-002"))
	_ = repo.Save(ctx, shipment)

	exists, err := repo.ExistsByReferenceNumber(ctx, "INT-002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected INT-002 to exist")
	}

	notExists, err := repo.ExistsByReferenceNumber(ctx, "INT-MISSING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notExists {
		t.Error("expected INT-MISSING not to exist")
	}
}

func TestIntegration_ShipmentRepo_Update(t *testing.T) {
	pool := setupDB(t)
	repo := postgres.NewShipmentRepo(pool)
	ctx := context.Background()

	shipment, _ := domain.NewShipment(testInput("INT-003"))
	_ = repo.Save(ctx, shipment)

	if err := shipment.Transition(domain.StatusPickedUp); err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	if err := repo.Update(ctx, shipment); err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	found, _ := repo.FindByID(ctx, shipment.ID)
	if found.CurrentStatus != domain.StatusPickedUp {
		t.Errorf("expected status PICKED_UP after update, got %s", found.CurrentStatus)
	}
}

func TestIntegration_ShipmentRepo_Update_NotFound(t *testing.T) {
	pool := setupDB(t)
	repo := postgres.NewShipmentRepo(pool)
	ctx := context.Background()

	shipment, _ := domain.NewShipment(testInput("INT-004"))
	// not saved — update should fail

	err := repo.Update(ctx, shipment)
	if !errors.Is(err, domain.ErrShipmentNotFound) {
		t.Errorf("expected ErrShipmentNotFound, got %v", err)
	}
}

// --- EventRepo tests ---

func TestIntegration_EventRepo_SaveAndFindByShipmentID(t *testing.T) {
	pool := setupDB(t)
	shipmentRepo := postgres.NewShipmentRepo(pool)
	eventRepo := postgres.NewEventRepo(pool)
	ctx := context.Background()

	shipment, _ := domain.NewShipment(testInput("INT-005"))
	_ = shipmentRepo.Save(ctx, shipment)

	e1 := &domain.ShipmentEvent{
		ID:         "evt-001",
		ShipmentID: shipment.ID,
		Status:     domain.StatusPending,
		Notes:      "created",
		OccurredAt: time.Now(),
	}
	e2 := &domain.ShipmentEvent{
		ID:         "evt-002",
		ShipmentID: shipment.ID,
		Status:     domain.StatusPickedUp,
		Notes:      "picked up",
		OccurredAt: time.Now().Add(time.Second),
	}

	_ = eventRepo.Save(ctx, e1)
	_ = eventRepo.Save(ctx, e2)

	events, err := eventRepo.FindByShipmentID(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("FindByShipmentID failed: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
	if events[0].Status != domain.StatusPending {
		t.Errorf("expected first event PENDING, got %s", events[0].Status)
	}
	if events[1].Status != domain.StatusPickedUp {
		t.Errorf("expected second event PICKED_UP, got %s", events[1].Status)
	}
}

func TestIntegration_EventRepo_FindByShipmentID_Empty(t *testing.T) {
	pool := setupDB(t)
	shipmentRepo := postgres.NewShipmentRepo(pool)
	eventRepo := postgres.NewEventRepo(pool)
	ctx := context.Background()

	shipment, _ := domain.NewShipment(testInput("INT-006"))
	_ = shipmentRepo.Save(ctx, shipment)

	events, err := eventRepo.FindByShipmentID(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events, got %d", len(events))
	}
}
