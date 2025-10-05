CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE TYPE artwork_status AS ENUM (
    'sold',
    'available',
    'coming_soon',
    'not_for_sale',
    'unavailable'
);
CREATE TYPE artwork_medium AS ENUM (
    'oil_panel',
    'acrylic_panel',
    'oil_mdf',
    'oil_paper',
    'unknown'
);
CREATE TYPE artwork_category AS ENUM ('figure', 'landscape', 'multi_figure', 'other');
CREATE TABLE artworks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    title VARCHAR(255) NOT NULL,
    painting_number INTEGER,
    painting_year INTEGER,
    width_inches DECIMAL(8, 4),
    height_inches DECIMAL(8, 4),
    price_cents INTEGER,
    paper BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    sold_at TIMESTAMP,
    status artwork_status NOT NULL,
    medium artwork_medium NOT NULL,
    category artwork_category NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);
CREATE TABLE images (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    artwork_id UUID REFERENCES artworks (id) ON DELETE CASCADE,
    is_main_image BOOLEAN DEFAULT FALSE,
    image_url TEXT NOT NULL,
    image_width INTEGER,
    image_height INTEGER,
    created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);
CREATE INDEX idx_artworks_status ON artworks (status);
CREATE INDEX idx_artworks_sort_order ON artworks (sort_order);
CREATE INDEX idx_artworks_created_at ON artworks (created_at DESC);
CREATE INDEX idx_artworks_status_sort_order ON artworks (status, sort_order);
CREATE INDEX idx_images_artwork_id ON images (artwork_id);
