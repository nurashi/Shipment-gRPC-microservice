CREATE TABLE IF NOT EXISTS shipments (
    id               UUID PRIMARY KEY,
    reference_number VARCHAR(255) UNIQUE NOT NULL,
    origin           VARCHAR(255) NOT NULL,
    destination      VARCHAR(255) NOT NULL,
    status           VARCHAR(50)  NOT NULL DEFAULT 'PENDING',
    driver_name      VARCHAR(255) NOT NULL DEFAULT '',
    unit_number      VARCHAR(255) NOT NULL DEFAULT '',
    amount           NUMERIC(12,2) NOT NULL DEFAULT 0,
    driver_revenue   NUMERIC(12,2) NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS shipment_events (
    id          UUID PRIMARY KEY,
    shipment_id UUID NOT NULL REFERENCES shipments(id) ON DELETE CASCADE,
    status      VARCHAR(50) NOT NULL,
    notes       TEXT NOT NULL DEFAULT '',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_shipment_events_shipment_id ON shipment_events(shipment_id);
CREATE INDEX IF NOT EXISTS idx_shipments_reference_number ON shipments(reference_number);
