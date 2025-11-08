-- Artworks indexes
DROP INDEX IF EXISTS idx_images_artwork_id;

DROP INDEX IF EXISTS idx_artworks_order_id;

DROP INDEX IF EXISTS idx_artworks_status_sort_order;

DROP INDEX IF EXISTS idx_artworks_created_at;

DROP INDEX IF EXISTS idx_artworks_sort_order;

DROP INDEX IF EXISTS idx_artworks_status;

-- Orders and payments indexes
DROP INDEX IF EXISTS idx_payments_stripe_payment_intent_id;

DROP INDEX IF EXISTS idx_payments_created_at;

DROP INDEX IF EXISTS idx_payments_status;

DROP INDEX IF EXISTS idx_payments_order_id;

DROP INDEX IF EXISTS idx_shipping_details_email;

DROP INDEX IF EXISTS idx_orders_stripe_session_id;

DROP INDEX IF EXISTS idx_orders_created_at;

DROP INDEX IF EXISTS idx_orders_status;

-- Auth indexes
DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;

DROP INDEX IF EXISTS idx_refresh_tokens_revoked;

DROP INDEX IF EXISTS idx_refresh_tokens_jti;

DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;

DROP INDEX IF EXISTS idx_refresh_tokens_user_id;

-- Artworks tables
DROP TABLE IF EXISTS images CASCADE;

DROP TABLE IF EXISTS artworks CASCADE;

-- Orders and payments tables
DROP TABLE IF EXISTS payments;

DROP TABLE IF EXISTS payment_requirements;

DROP TABLE IF EXISTS shipping_details;

DROP TABLE IF EXISTS orders;

-- Auth tables
DROP TABLE IF EXISTS refresh_tokens;

DROP TABLE IF EXISTS users;

-- Types
DROP TYPE IF EXISTS payment_status;

DROP TYPE IF EXISTS order_status;

DROP TYPE IF EXISTS artwork_category;

DROP TYPE IF EXISTS artwork_medium;

DROP TYPE IF EXISTS artwork_status;