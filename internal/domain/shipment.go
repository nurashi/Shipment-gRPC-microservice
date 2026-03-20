package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Shipment struct {
	ID              string
	ReferenceNumber string
	Origin          string
	Destination     string
	CurrentStatus   Status
	DriverName      string
	UnitNumber      string
	Amount          float64
	DriverRevenue   float64
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CreateShipmentInput struct {
	ReferenceNumber string
	Origin          string
	Destination     string
	DriverName      string
	UnitNumber      string
	Amount          float64
	DriverRevenue   float64
}

func NewShipment(input CreateShipmentInput) (*Shipment, error) {
	if input.ReferenceNumber == "" {
		return nil, fmt.Errorf("%w: reference_number", ErrMissingRequiredField)
	}
	if input.Origin == "" {
		return nil, fmt.Errorf("%w: origin", ErrMissingRequiredField)
	}
	if input.Destination == "" {
		return nil, fmt.Errorf("%w: destination", ErrMissingRequiredField)
	}
	if input.DriverName == "" {
		return nil, fmt.Errorf("%w: driver_name", ErrMissingRequiredField)
	}
	if input.UnitNumber == "" {
		return nil, fmt.Errorf("%w: unit_number", ErrMissingRequiredField)
	}
	if input.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be greater than zero", ErrInvalidFieldValue)
	}
	if input.DriverRevenue < 0 {
		return nil, fmt.Errorf("%w: driver_revenue must be non-negative", ErrInvalidFieldValue)
	}

	now := time.Now()
	return &Shipment{
		ID:              uuid.New().String(),
		ReferenceNumber: input.ReferenceNumber,
		Origin:          input.Origin,
		Destination:     input.Destination,
		CurrentStatus:   StatusPending,
		DriverName:      input.DriverName,
		UnitNumber:      input.UnitNumber,
		Amount:          input.Amount,
		DriverRevenue:   input.DriverRevenue,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (s *Shipment) Transition(next Status) error {
	if s.CurrentStatus.IsTerminal() {
		return fmt.Errorf(
			"%w: cannot transition from %s",
			ErrShipmentTerminated, s.CurrentStatus,
		)
	}
	if !s.CurrentStatus.CanTransitionTo(next) {
		return fmt.Errorf(
			"%w: %s -> %s",
			ErrInvalidStatusTransition, s.CurrentStatus, next,
		)
	}

	s.CurrentStatus = next
	s.UpdatedAt = time.Now()
	return nil
}
