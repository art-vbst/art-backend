-- Create ENUM types for status fields
CREATE TYPE order_status AS ENUM (
    'pending',
    'processing',
    'shipped',
    'completed',
    'failed',
    'refunded'
);

CREATE TYPE payment_status AS ENUM ('succeeded', 'failed', 'refunded');

-- Create orders table
CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stripe_session_id VARCHAR(200) UNIQUE,
    stripe_payment_intent_id VARCHAR(200) UNIQUE,
    customer_email VARCHAR(255) NOT NULL,
    shipping_rate_id VARCHAR(200) NOT NULL,
    shipping_name VARCHAR(200) NOT NULL,
    shipping_address_line1 VARCHAR(200) NOT NULL,
    shipping_address_line2 VARCHAR(200),
    shipping_city VARCHAR(200) NOT NULL,
    shipping_postal_code VARCHAR(200) NOT NULL,
    shipping_state VARCHAR(200) NOT NULL,
    shipping_country VARCHAR(200) NOT NULL,
    subtotal_cents INTEGER NOT NULL,
    shipping_cents INTEGER NOT NULL,
    total_cents INTEGER NOT NULL,
    currency VARCHAR(200) NOT NULL,
    status order_status NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

-- Create payments table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    order_id UUID NOT NULL UNIQUE REFERENCES orders (id) ON DELETE CASCADE,
    stripe_payment_intent_id VARCHAR(200) NOT NULL,
    subtotal_cents INTEGER NOT NULL,
    shipping_cents INTEGER NOT NULL,
    shipping_stripe_id VARCHAR(200) NOT NULL,
    total_cents INTEGER NOT NULL,
    currency VARCHAR(200) NOT NULL,
    status payment_status NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

-- Create indexes for orders table
CREATE INDEX idx_orders_status ON orders (status);

CREATE INDEX idx_orders_created_at ON orders (created_at DESC);

CREATE INDEX idx_orders_customer_email ON orders (customer_email);

CREATE INDEX idx_orders_stripe_session_id ON orders (stripe_session_id);

CREATE INDEX idx_orders_stripe_payment_intent_id ON orders (stripe_payment_intent_id);

-- Create indexes for payments table
CREATE INDEX idx_payments_order_id ON payments (order_id);

CREATE INDEX idx_payments_status ON payments (status);

CREATE INDEX idx_payments_created_at ON payments (created_at DESC);

CREATE INDEX idx_payments_stripe_payment_intent_id ON payments (stripe_payment_intent_id);