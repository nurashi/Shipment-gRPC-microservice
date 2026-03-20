package ports

import (
	"context"

	"github.com/nurashi/Shipment-gRPC-microservice/internal/domain"
)

type ShipmentRepository interface {
	Save(ctx context.Context, shipment *domain.Shipment) error
	FindByID(ctx context.Context, id string) (*domain.Shipment, error)
	ExistsByReferenceNumber(ctx context.Context, referenceNumber string) (bool, error)
	Update(ctx context.Context, shipment *domain.Shipment) error
}

type ShipmentEventRepository interface {
	Save(ctx context.Context, event *domain.ShipmentEvent) error
	FindByShipmentID(ctx context.Context, shipmentID string) ([]*domain.ShipmentEvent, error)
}
