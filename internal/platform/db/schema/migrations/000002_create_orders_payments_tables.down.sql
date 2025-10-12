DROP INDEX IF EXISTS idx_payments_stripe_payment_intent_id;

DROP INDEX IF EXISTS idx_payments_created_at;

DROP INDEX IF EXISTS idx_payments_status;

DROP INDEX IF EXISTS idx_payments_order_id;

DROP INDEX IF EXISTS idx_shipping_details_email;

DROP INDEX IF EXISTS idx_orders_stripe_session_id;

DROP INDEX IF EXISTS idx_orders_created_at;

DROP INDEX IF EXISTS idx_orders_status;

DROP INDEX IF EXISTS idx_artworks_order_id;

ALTER TABLE artworks DROP COLUMN IF EXISTS order_id;

DROP TABLE IF EXISTS payments;

DROP TABLE IF EXISTS payment_requirements;

DROP TABLE IF EXISTS shipping_details;

DROP TABLE IF EXISTS orders;

DROP TYPE IF EXISTS payment_status;

DROP TYPE IF EXISTS order_status;