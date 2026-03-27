-- Create owner_settlements table
CREATE TABLE IF NOT EXISTS owner_settlements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    booking_id UUID NOT NULL,
    owner_id UUID NOT NULL,

    gross_amount NUMERIC(10,2) NOT NULL,
    platform_fee NUMERIC(10,2) NOT NULL DEFAULT 0,
    discount_amount NUMERIC(10,2) NOT NULL DEFAULT 0,
    net_amount NUMERIC(10,2) NOT NULL,

    status VARCHAR(30) NOT NULL DEFAULT 'pending',
    -- pending / available / processing / paid / hold / reversed

    available_at TIMESTAMP NULL,
    paid_at TIMESTAMP NULL,

    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT owner_settlements_booking_id_fkey
        FOREIGN KEY (booking_id) REFERENCES bookings(id),
    CONSTRAINT owner_settlements_owner_id_fkey
        FOREIGN KEY (owner_id) REFERENCES users(id),
    CONSTRAINT owner_settlements_booking_unique UNIQUE (booking_id)
);
