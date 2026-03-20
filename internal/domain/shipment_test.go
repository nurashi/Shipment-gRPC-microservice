package domain_test

import (
	"testing"

	"github.com/nurashi/Shipment-gRPC-microservice/internal/domain"
)

func validInput() domain.CreateShipmentInput {
	return domain.CreateShipmentInput{
		ReferenceNumber: "REF-001",
		Origin:          "Almaty",
		Destination:     "Astana",
		DriverName:      "Nurassyl Orazbek",
		UnitNumber:      "TRUCK-42",
		Amount:          1500.00,
		DriverRevenue:   500.00,
	}
}

func TestNewShipment_Success(t *testing.T) {
	s, err := domain.NewShipment(validInput())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if s.CurrentStatus != domain.StatusPending {
		t.Fatalf("expected status %s, got %s", domain.StatusPending, s.CurrentStatus)
	}
	if s.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if s.ReferenceNumber != "REF-001" {
		t.Fatalf("expected reference number REF-001, got %s", s.ReferenceNumber)
	}
}

func TestNewShipment_MissingReferenceNumber(t *testing.T) {
	input := validInput()
	input.ReferenceNumber = ""
	_, err := domain.NewShipment(input)
	if err == nil {
		t.Fatal("expected error for missing reference number")
	}
}

func TestNewShipment_MissingOrigin(t *testing.T) {
	input := validInput()
	input.Origin = ""
	_, err := domain.NewShipment(input)
	if err == nil {
		t.Fatal("expected error for missing origin")
	}
}

func TestNewShipment_MissingDestination(t *testing.T) {
	input := validInput()
	input.Destination = ""
	_, err := domain.NewShipment(input)
	if err == nil {
		t.Fatal("expected error for missing destination")
	}
}

func TestNewShipment_MissingDriverName(t *testing.T) {
	input := validInput()
	input.DriverName = ""
	_, err := domain.NewShipment(input)
	if err == nil {
		t.Fatal("expected error for missing driver_name")
	}
}

func TestNewShipment_MissingUnitNumber(t *testing.T) {
	input := validInput()
	input.UnitNumber = ""
	_, err := domain.NewShipment(input)
	if err == nil {
		t.Fatal("expected error for missing unit_number")
	}
}

func TestNewShipment_ZeroAmount(t *testing.T) {
	input := validInput()
	input.Amount = 0
	_, err := domain.NewShipment(input)
	if err == nil {
		t.Fatal("expected error for zero amount")
	}
}

func TestNewShipment_NegativeAmount(t *testing.T) {
	input := validInput()
	input.Amount = -100
	_, err := domain.NewShipment(input)
	if err == nil {
		t.Fatal("expected error for negative amount")
	}
}

func TestNewShipment_NegativeDriverRevenue(t *testing.T) {
	input := validInput()
	input.DriverRevenue = -1
	_, err := domain.NewShipment(input)
	if err == nil {
		t.Fatal("expected error for negative driver_revenue")
	}
}

func TestNewShipment_ZeroDriverRevenue_Allowed(t *testing.T) {
	input := validInput()
	input.DriverRevenue = 0
	_, err := domain.NewShipment(input)
	if err != nil {
		t.Fatalf("expected zero driver_revenue to be allowed, got: %v", err)
	}
}

func TestTransition_FullHappyPath(t *testing.T) {
	s, _ := domain.NewShipment(validInput())

	transitions := []domain.Status{
		domain.StatusPickedUp,
		domain.StatusInTransit,
		domain.StatusDelivered,
	}

	for _, next := range transitions {
		if err := s.Transition(next); err != nil {
			t.Fatalf("transition to %s failed: %v", next, err)
		}
		if s.CurrentStatus != next {
			t.Fatalf("expected status %s, got %s", next, s.CurrentStatus)
		}
	}
}

func TestTransition_PendingToDelivered_Rejected(t *testing.T) {
	s, _ := domain.NewShipment(validInput())

	err := s.Transition(domain.StatusDelivered)
	if err == nil {
		t.Fatal("expected error for invalid transition PENDING -> DELIVERED")
	}
}

func TestTransition_PendingToInTransit_Rejected(t *testing.T) {
	s, _ := domain.NewShipment(validInput())

	err := s.Transition(domain.StatusInTransit)
	if err == nil {
		t.Fatal("expected error for invalid transition PENDING -> IN_TRANSIT")
	}
}

func TestTransition_FromTerminal_Delivered(t *testing.T) {
	s, _ := domain.NewShipment(validInput())
	_ = s.Transition(domain.StatusPickedUp)
	_ = s.Transition(domain.StatusInTransit)
	_ = s.Transition(domain.StatusDelivered)

	err := s.Transition(domain.StatusPending)
	if err == nil {
		t.Fatal("expected error when transitioning from terminal state DELIVERED")
	}
}

func TestTransition_FromTerminal_Cancelled(t *testing.T) {
	s, _ := domain.NewShipment(validInput())
	_ = s.Transition(domain.StatusCancelled)

	err := s.Transition(domain.StatusPickedUp)
	if err == nil {
		t.Fatal("expected error when transitioning from terminal state CANCELLED")
	}
}

func TestTransition_CancelFromAnyNonTerminal(t *testing.T) {
	cases := []struct {
		name    string
		setup   []domain.Status
		current domain.Status
	}{
		{"from PENDING", nil, domain.StatusPending},
		{"from PICKED_UP", []domain.Status{domain.StatusPickedUp}, domain.StatusPickedUp},
		{"from IN_TRANSIT", []domain.Status{domain.StatusPickedUp, domain.StatusInTransit}, domain.StatusInTransit},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, _ := domain.NewShipment(validInput())
			for _, st := range tc.setup {
				_ = s.Transition(st)
			}
			if err := s.Transition(domain.StatusCancelled); err != nil {
				t.Fatalf("expected cancel from %s to succeed, got %v", tc.current, err)
			}
		})
	}
}

func TestTransition_DuplicateStatus_Rejected(t *testing.T) {
	s, _ := domain.NewShipment(validInput())
	_ = s.Transition(domain.StatusPickedUp)

	err := s.Transition(domain.StatusPickedUp)
	if err == nil {
		t.Fatal("expected error for duplicate status transition")
	}
}
