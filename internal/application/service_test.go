package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nurashi/Shipment-gRPC-microservice/internal/application"
	"github.com/nurashi/Shipment-gRPC-microservice/internal/domain"
)

// --- Mock implementations ---

type mockShipmentRepo struct {
	shipments map[string]*domain.Shipment
	refs      map[string]bool
}

func newMockShipmentRepo() *mockShipmentRepo {
	return &mockShipmentRepo{
		shipments: make(map[string]*domain.Shipment),
		refs:      make(map[string]bool),
	}
}

func (m *mockShipmentRepo) Save(_ context.Context, s *domain.Shipment) error {
	m.shipments[s.ID] = s
	m.refs[s.ReferenceNumber] = true
	return nil
}

func (m *mockShipmentRepo) FindByID(_ context.Context, id string) (*domain.Shipment, error) {
	s, ok := m.shipments[id]
	if !ok {
		return nil, domain.ErrShipmentNotFound
	}
	return s, nil
}

func (m *mockShipmentRepo) ExistsByReferenceNumber(_ context.Context, ref string) (bool, error) {
	return m.refs[ref], nil
}

func (m *mockShipmentRepo) Update(_ context.Context, s *domain.Shipment) error {
	if _, ok := m.shipments[s.ID]; !ok {
		return domain.ErrShipmentNotFound
	}
	m.shipments[s.ID] = s
	return nil
}

type mockEventRepo struct {
	events map[string][]*domain.ShipmentEvent
}

func newMockEventRepo() *mockEventRepo {
	return &mockEventRepo{
		events: make(map[string][]*domain.ShipmentEvent),
	}
}

func (m *mockEventRepo) Save(_ context.Context, e *domain.ShipmentEvent) error {
	m.events[e.ShipmentID] = append(m.events[e.ShipmentID], e)
	return nil
}

func (m *mockEventRepo) FindByShipmentID(_ context.Context, shipmentID string) ([]*domain.ShipmentEvent, error) {
	return m.events[shipmentID], nil
}

// --- Helpers ---

func newTestService() (application.ShipmentService, *mockShipmentRepo, *mockEventRepo) {
	sr := newMockShipmentRepo()
	er := newMockEventRepo()
	svc := application.NewShipmentService(sr, er)
	return svc, sr, er
}

func validInput() domain.CreateShipmentInput {
	return domain.CreateShipmentInput{
		ReferenceNumber: "REF-001",
		Origin:          "Almaty",
		Destination:     "Astana",
		DriverName:      "John Doe",
		UnitNumber:      "TRUCK-42",
		Amount:          1500.00,
		DriverRevenue:   500.00,
	}
}

// --- Tests ---

func TestCreateShipment_Success(t *testing.T) {
	svc, sr, er := newTestService()
	ctx := context.Background()

	shipment, err := svc.CreateShipment(ctx, validInput())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if shipment.CurrentStatus != domain.StatusPending {
		t.Fatalf("expected status PENDING, got %s", shipment.CurrentStatus)
	}

	if _, ok := sr.shipments[shipment.ID]; !ok {
		t.Fatal("shipment not saved to repository")
	}

	events := er.events[shipment.ID]
	if len(events) != 1 {
		t.Fatalf("expected 1 initial event, got %d", len(events))
	}
	if events[0].Status != domain.StatusPending {
		t.Fatalf("expected initial event status PENDING, got %s", events[0].Status)
	}
}

func TestCreateShipment_DuplicateReferenceNumber(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, _ = svc.CreateShipment(ctx, validInput())

	_, err := svc.CreateShipment(ctx, validInput())
	if !errors.Is(err, domain.ErrDuplicateReferenceNumber) {
		t.Fatalf("expected ErrDuplicateReferenceNumber, got %v", err)
	}
}

func TestCreateShipment_MissingRequiredField(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	input := validInput()
	input.ReferenceNumber = ""

	_, err := svc.CreateShipment(ctx, input)
	if !errors.Is(err, domain.ErrMissingRequiredField) {
		t.Fatalf("expected ErrMissingRequiredField, got %v", err)
	}
}

func TestGetShipment_Success(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	created, _ := svc.CreateShipment(ctx, validInput())

	found, err := svc.GetShipment(ctx, created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.ID != created.ID {
		t.Fatalf("expected ID %s, got %s", created.ID, found.ID)
	}
}

func TestGetShipment_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.GetShipment(ctx, "non-existent-id")
	if !errors.Is(err, domain.ErrShipmentNotFound) {
		t.Fatalf("expected ErrShipmentNotFound, got %v", err)
	}
}

func TestAddShipmentEvent_ValidTransition(t *testing.T) {
	svc, sr, er := newTestService()
	ctx := context.Background()

	shipment, _ := svc.CreateShipment(ctx, validInput())

	event, err := svc.AddShipmentEvent(ctx, shipment.ID, domain.StatusPickedUp, "driver picked up cargo")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if event.Status != domain.StatusPickedUp {
		t.Fatalf("expected event status PICKED_UP, got %s", event.Status)
	}

	updated := sr.shipments[shipment.ID]
	if updated.CurrentStatus != domain.StatusPickedUp {
		t.Fatalf("expected shipment status PICKED_UP, got %s", updated.CurrentStatus)
	}

	events := er.events[shipment.ID]
	if len(events) != 2 {
		t.Fatalf("expected 2 events (initial + transition), got %d", len(events))
	}
}

func TestAddShipmentEvent_InvalidTransition(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	shipment, _ := svc.CreateShipment(ctx, validInput())

	_, err := svc.AddShipmentEvent(ctx, shipment.ID, domain.StatusDelivered, "skip to delivered")
	if !errors.Is(err, domain.ErrInvalidStatusTransition) {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
}

func TestAddShipmentEvent_TerminalState(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	shipment, _ := svc.CreateShipment(ctx, validInput())
	_, _ = svc.AddShipmentEvent(ctx, shipment.ID, domain.StatusCancelled, "cancelled")

	_, err := svc.AddShipmentEvent(ctx, shipment.ID, domain.StatusPickedUp, "try to pick up")
	if !errors.Is(err, domain.ErrShipmentTerminated) {
		t.Fatalf("expected ErrShipmentTerminated, got %v", err)
	}
}

func TestAddShipmentEvent_ShipmentNotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.AddShipmentEvent(ctx, "non-existent-id", domain.StatusPickedUp, "")
	if !errors.Is(err, domain.ErrShipmentNotFound) {
		t.Fatalf("expected ErrShipmentNotFound, got %v", err)
	}
}

func TestGetShipmentHistory_Success(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	shipment, _ := svc.CreateShipment(ctx, validInput())
	_, _ = svc.AddShipmentEvent(ctx, shipment.ID, domain.StatusPickedUp, "picked up")
	_, _ = svc.AddShipmentEvent(ctx, shipment.ID, domain.StatusInTransit, "in transit")

	events, err := svc.GetShipmentHistory(ctx, shipment.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestGetShipmentHistory_ShipmentNotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	_, err := svc.GetShipmentHistory(ctx, "non-existent-id")
	if !errors.Is(err, domain.ErrShipmentNotFound) {
		t.Fatalf("expected ErrShipmentNotFound, got %v", err)
	}
}
