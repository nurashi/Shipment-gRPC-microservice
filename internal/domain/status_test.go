package domain_test

import (
	"testing"

	"github.com/nurashi/Shipment-gRPC-microservice/internal/domain"
)

func TestStatus_IsValid(t *testing.T) {
	valid := []domain.Status{
		domain.StatusPending,
		domain.StatusPickedUp,
		domain.StatusInTransit,
		domain.StatusDelivered,
		domain.StatusCancelled,
	}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("expected %s to be valid", s)
		}
	}

	if domain.Status("UNKNOWN").IsValid() {
		t.Error("expected UNKNOWN to be invalid")
	}
	if domain.Status("").IsValid() {
		t.Error("expected empty string to be invalid")
	}
}

func TestStatus_IsTerminal(t *testing.T) {
	terminal := []domain.Status{domain.StatusDelivered, domain.StatusCancelled}
	for _, s := range terminal {
		if !s.IsTerminal() {
			t.Errorf("expected %s to be terminal", s)
		}
	}

	nonTerminal := []domain.Status{domain.StatusPending, domain.StatusPickedUp, domain.StatusInTransit}
	for _, s := range nonTerminal {
		if s.IsTerminal() {
			t.Errorf("expected %s to be non-terminal", s)
		}
	}
}

func TestStatus_CanTransitionTo(t *testing.T) {
	cases := []struct {
		from    domain.Status
		to      domain.Status
		allowed bool
	}{
		{domain.StatusPending, domain.StatusPickedUp, true},
		{domain.StatusPending, domain.StatusCancelled, true},
		{domain.StatusPending, domain.StatusInTransit, false},
		{domain.StatusPending, domain.StatusDelivered, false},
		{domain.StatusPending, domain.StatusPending, false},
		{domain.StatusPickedUp, domain.StatusInTransit, true},
		{domain.StatusPickedUp, domain.StatusCancelled, true},
		{domain.StatusPickedUp, domain.StatusDelivered, false},
		{domain.StatusPickedUp, domain.StatusPending, false},
		{domain.StatusInTransit, domain.StatusDelivered, true},
		{domain.StatusInTransit, domain.StatusCancelled, true},
		{domain.StatusInTransit, domain.StatusPickedUp, false},
		{domain.StatusDelivered, domain.StatusCancelled, false},
		{domain.StatusDelivered, domain.StatusPending, false},
		{domain.StatusCancelled, domain.StatusPending, false},
		{domain.StatusCancelled, domain.StatusPickedUp, false},
	}

	for _, tc := range cases {
		got := tc.from.CanTransitionTo(tc.to)
		if got != tc.allowed {
			t.Errorf("%s -> %s: expected allowed=%v, got %v", tc.from, tc.to, tc.allowed, got)
		}
	}
}

func TestNewShipmentEvent_FieldsPopulated(t *testing.T) {
	event := domain.NewShipmentEvent("ship-123", domain.StatusPickedUp, "picked up at depot")

	if event.ID == "" {
		t.Error("expected non-empty ID")
	}
	if event.ShipmentID != "ship-123" {
		t.Errorf("expected shipment ID ship-123, got %s", event.ShipmentID)
	}
	if event.Status != domain.StatusPickedUp {
		t.Errorf("expected status PICKED_UP, got %s", event.Status)
	}
	if event.Notes != "picked up at depot" {
		t.Errorf("expected notes 'picked up at depot', got %s", event.Notes)
	}
	if event.OccurredAt.IsZero() {
		t.Error("expected OccurredAt to be set")
	}
}

func TestNewShipmentEvent_UniqueIDs(t *testing.T) {
	e1 := domain.NewShipmentEvent("ship-1", domain.StatusPending, "")
	e2 := domain.NewShipmentEvent("ship-1", domain.StatusPending, "")

	if e1.ID == e2.ID {
		t.Error("expected two events to have different IDs")
	}
}
