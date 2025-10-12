CREATE TYPE order_status AS ENUM (
    'pending',
    'processing',
    'shipped',
    'completed',
    'failed',
    'canceled'
);

CREATE TYPE payment_status AS ENUM ('success', 'failed', 'refunded');

CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    status order_status NOT NULL,
    stripe_session_id TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE TABLE shipping_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL UNIQUE REFERENCES orders (id) ON DELETE CASCADE,
    email TEXT NOT NULL,
    name TEXT NOT NULL,
    line1 TEXT NOT NULL,
    line2 TEXT,
    city TEXT NOT NULL,
    state TEXT NOT NULL,
    postal TEXT NOT NULL,
    country TEXT NOT NULL
);

CREATE TABLE payment_requirements (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL UNIQUE REFERENCES orders (id) ON DELETE CASCADE,
    subtotal_cents INTEGER NOT NULL,
    shipping_cents INTEGER NOT NULL,
    total_cents INTEGER NOT NULL,
    currency TEXT NOT NULL
);

CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    stripe_payment_intent_id TEXT NOT NULL,
    status payment_status NOT NULL,
    total_cents INTEGER NOT NULL,
    currency TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    paid_at TIMESTAMP
);

ALTER TABLE artworks
ADD COLUMN order_id UUID REFERENCES orders (id) ON DELETE
SET NULL;

CREATE INDEX idx_artworks_order_id ON artworks (order_id);

CREATE INDEX idx_orders_status ON orders (status);

CREATE INDEX idx_orders_created_at ON orders (created_at);

CREATE INDEX idx_orders_stripe_session_id ON orders (stripe_session_id);

CREATE INDEX idx_shipping_details_email ON shipping_details (email);

CREATE INDEX idx_payments_order_id ON payments (order_id);

CREATE INDEX idx_payments_status ON payments (status);

CREATE INDEX idx_payments_created_at ON payments (created_at);

CREATE INDEX idx_payments_stripe_payment_intent_id ON payments (stripe_payment_intent_id);