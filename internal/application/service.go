package application

import (
	"context"

	"github.com/nurashi/Shipment-gRPC-microservice/internal/domain"
	"github.com/nurashi/Shipment-gRPC-microservice/internal/ports"
)

type ShipmentService interface {
	CreateShipment(ctx context.Context, input domain.CreateShipmentInput) (*domain.Shipment, error)
	GetShipment(ctx context.Context, id string) (*domain.Shipment, error)
	AddShipmentEvent(ctx context.Context, shipmentID string, status domain.Status, notes string) (*domain.ShipmentEvent, error)
	GetShipmentHistory(ctx context.Context, shipmentID string) ([]*domain.ShipmentEvent, error)
}

type shipmentService struct {
	shipments ports.ShipmentRepository
	events    ports.ShipmentEventRepository
}

func NewShipmentService(
	shipments ports.ShipmentRepository,
	events ports.ShipmentEventRepository,
) ShipmentService {
	return &shipmentService{
		shipments: shipments,
		events:    events,
	}
}

func (s *shipmentService) CreateShipment(ctx context.Context, input domain.CreateShipmentInput) (*domain.Shipment, error) {
	exists, err := s.shipments.ExistsByReferenceNumber(ctx, input.ReferenceNumber)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrDuplicateReferenceNumber
	}

	shipment, err := domain.NewShipment(input)
	if err != nil {
		return nil, err
	}

	if err := s.shipments.Save(ctx, shipment); err != nil {
		return nil, err
	}

	initialEvent := domain.NewShipmentEvent(shipment.ID, domain.StatusPending, "shipment created")
	if err := s.events.Save(ctx, initialEvent); err != nil {
		return nil, err
	}

	return shipment, nil
}

func (s *shipmentService) GetShipment(ctx context.Context, id string) (*domain.Shipment, error) {
	return s.shipments.FindByID(ctx, id)
}

func (s *shipmentService) AddShipmentEvent(ctx context.Context, shipmentID string, status domain.Status, notes string) (*domain.ShipmentEvent, error) {
	if !status.IsValid() {
		return nil, domain.ErrInvalidStatusTransition
	}

	shipment, err := s.shipments.FindByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	if err := shipment.Transition(status); err != nil {
		return nil, err
	}

	if err := s.shipments.Update(ctx, shipment); err != nil {
		return nil, err
	}

	event := domain.NewShipmentEvent(shipmentID, status, notes)
	if err := s.events.Save(ctx, event); err != nil {
		return nil, err
	}

	return event, nil
}

func (s *shipmentService) GetShipmentHistory(ctx context.Context, shipmentID string) ([]*domain.ShipmentEvent, error) {
	_, err := s.shipments.FindByID(ctx, shipmentID)
	if err != nil {
		return nil, err
	}

	return s.events.FindByShipmentID(ctx, shipmentID)
}
