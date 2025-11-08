CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Types
CREATE TYPE artwork_status AS ENUM (
    'available',
    'sold',
    'not_for_sale',
    'unavailable',
    'coming_soon'
);

CREATE TYPE artwork_medium AS ENUM (
    'oil_panel',
    'acrylic_panel',
    'oil_mdf',
    'oil_paper',
    'unknown'
);

CREATE TYPE artwork_category AS ENUM ('figure', 'landscape', 'multi_figure', 'other');

CREATE TYPE order_status AS ENUM (
    'pending',
    'processing',
    'shipped',
    'completed',
    'failed',
    'canceled'
);

CREATE TYPE payment_status AS ENUM ('success', 'failed', 'refunded');

-- Auth tables
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash TEXT NOT NULL,
    jti UUID UNIQUE NOT NULL,
    session_id UUID NOT NULL,
    created_at TIMESTAMP DEFAULT now(),
    expires_at TIMESTAMP NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE
);

-- Orders and payments tables
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

-- Artworks tables
CREATE TABLE artworks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    painting_number INTEGER,
    painting_year INTEGER,
    width_inches DECIMAL(8, 4) NOT NULL,
    height_inches DECIMAL(8, 4) NOT NULL,
    price_cents INTEGER NOT NULL,
    paper BOOLEAN DEFAULT FALSE,
    sort_order INTEGER NOT NULL DEFAULT 0,
    sold_at TIMESTAMP,
    status artwork_status NOT NULL,
    medium artwork_medium NOT NULL,
    category artwork_category NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    order_id UUID REFERENCES orders (id) ON DELETE
    SET NULL
);

CREATE TABLE images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    artwork_id UUID REFERENCES artworks (id) ON DELETE CASCADE,
    is_main_image BOOLEAN NOT NULL DEFAULT FALSE,
    object_name TEXT NOT NULL,
    image_url TEXT NOT NULL,
    image_width INTEGER,
    image_height INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

-- Auth indexes
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens (token_hash);

CREATE INDEX idx_refresh_tokens_jti ON refresh_tokens (jti);

CREATE INDEX idx_refresh_tokens_revoked ON refresh_tokens (revoked);

CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

-- Orders and payments indexes
CREATE INDEX idx_orders_status ON orders (status);

CREATE INDEX idx_orders_created_at ON orders (created_at);

CREATE INDEX idx_orders_stripe_session_id ON orders (stripe_session_id);

CREATE INDEX idx_shipping_details_email ON shipping_details (email);

CREATE INDEX idx_payments_order_id ON payments (order_id);

CREATE INDEX idx_payments_status ON payments (status);

CREATE INDEX idx_payments_created_at ON payments (created_at);

CREATE INDEX idx_payments_stripe_payment_intent_id ON payments (stripe_payment_intent_id);

-- Artworks indexes
CREATE INDEX idx_artworks_status ON artworks (status);

CREATE INDEX idx_artworks_sort_order ON artworks (sort_order);

CREATE INDEX idx_artworks_created_at ON artworks (created_at DESC);

CREATE INDEX idx_artworks_status_sort_order ON artworks (status, sort_order);

CREATE INDEX idx_artworks_order_id ON artworks (order_id);

CREATE INDEX idx_images_artwork_id ON images (artwork_id);