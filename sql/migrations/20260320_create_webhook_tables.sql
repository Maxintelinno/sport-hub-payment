-- Create payment_webhook_logs table
CREATE TABLE IF NOT EXISTS payment_webhook_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    event_type VARCHAR(100) NULL,
    event_id VARCHAR(150) NULL,
    signature TEXT NULL,
    request_headers JSONB NULL,
    request_body JSONB NOT NULL,
    received_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP NULL,
    process_status VARCHAR(30) NOT NULL DEFAULT 'received',
    process_error TEXT NULL
);

-- Create payment_events table
CREATE TABLE IF NOT EXISTS payment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider VARCHAR(50) NOT NULL,
    event_id VARCHAR(150) NOT NULL,
    payment_id UUID NULL REFERENCES payments(id) ON DELETE SET NULL,
    booking_id UUID NULL REFERENCES bookings(id) ON DELETE SET NULL,
    event_type VARCHAR(100) NOT NULL,
    event_status VARCHAR(50) NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(provider, event_id)
);
