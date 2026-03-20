package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/nurashi/Shipment-gRPC-microservice/internal/domain"
)

type ShipmentRepo struct {
	pool *pgxpool.Pool
}

func NewShipmentRepo(pool *pgxpool.Pool) *ShipmentRepo {
	return &ShipmentRepo{pool: pool}
}

func (r *ShipmentRepo) Save(ctx context.Context, s *domain.Shipment) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO shipments (id, reference_number, origin, destination, status, driver_name, unit_number, amount, driver_revenue, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`, s.ID, s.ReferenceNumber, s.Origin, s.Destination, string(s.CurrentStatus),
		s.DriverName, s.UnitNumber, s.Amount, s.DriverRevenue, s.CreatedAt, s.UpdatedAt)
	return err
}

func (r *ShipmentRepo) FindByID(ctx context.Context, id string) (*domain.Shipment, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, reference_number, origin, destination, status, driver_name, unit_number, amount, driver_revenue, created_at, updated_at
		FROM shipments WHERE id = $1
	`, id)

	s := &domain.Shipment{}
	var status string
	err := row.Scan(
		&s.ID, &s.ReferenceNumber, &s.Origin, &s.Destination,
		&status, &s.DriverName, &s.UnitNumber,
		&s.Amount, &s.DriverRevenue, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrShipmentNotFound
	}
	if err != nil {
		return nil, err
	}
	s.CurrentStatus = domain.Status(status)
	return s, nil
}

func (r *ShipmentRepo) ExistsByReferenceNumber(ctx context.Context, ref string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM shipments WHERE reference_number = $1)`, ref,
	).Scan(&exists)
	return exists, err
}

func (r *ShipmentRepo) Update(ctx context.Context, s *domain.Shipment) error {
	tag, err := r.pool.Exec(ctx, `
		UPDATE shipments SET status = $1, driver_name = $2, unit_number = $3, amount = $4, driver_revenue = $5, updated_at = $6
		WHERE id = $7
	`, string(s.CurrentStatus), s.DriverName, s.UnitNumber, s.Amount, s.DriverRevenue, s.UpdatedAt, s.ID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrShipmentNotFound
	}
	return nil
}

type EventRepo struct {
	pool *pgxpool.Pool
}

func NewEventRepo(pool *pgxpool.Pool) *EventRepo {
	return &EventRepo{pool: pool}
}

func (r *EventRepo) Save(ctx context.Context, e *domain.ShipmentEvent) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO shipment_events (id, shipment_id, status, notes, occurred_at)
		VALUES ($1, $2, $3, $4, $5)
	`, e.ID, e.ShipmentID, string(e.Status), e.Notes, e.OccurredAt)
	return err
}

func (r *EventRepo) FindByShipmentID(ctx context.Context, shipmentID string) ([]*domain.ShipmentEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, shipment_id, status, notes, occurred_at
		FROM shipment_events WHERE shipment_id = $1
		ORDER BY occurred_at ASC
	`, shipmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []*domain.ShipmentEvent
	for rows.Next() {
		e := &domain.ShipmentEvent{}
		var status string
		if err := rows.Scan(&e.ID, &e.ShipmentID, &status, &e.Notes, &e.OccurredAt); err != nil {
			return nil, err
		}
		e.Status = domain.Status(status)
		events = append(events, e)
	}
	return events, rows.Err()
}
