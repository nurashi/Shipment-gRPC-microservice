package domain

import (
	"time"

	"github.com/google/uuid"
)

type ShipmentEvent struct {
	ID         string
	ShipmentID string
	Status     Status
	Notes      string
	OccurredAt time.Time
}

func NewShipmentEvent(shipmentID string, status Status, notes string) *ShipmentEvent {
	return &ShipmentEvent{
		ID:         uuid.New().String(),
		ShipmentID: shipmentID,
		Status:     status,
		Notes:      notes,
		OccurredAt: time.Now(),
	}
}
