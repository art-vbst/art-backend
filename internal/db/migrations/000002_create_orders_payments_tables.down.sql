-- Drop indexes for payments table
DROP INDEX IF EXISTS idx_payments_stripe_payment_intent_id;

DROP INDEX IF EXISTS idx_payments_created_at;

DROP INDEX IF EXISTS idx_payments_status;

DROP INDEX IF EXISTS idx_payments_order_id;

-- Drop indexes for orders table
DROP INDEX IF EXISTS idx_orders_stripe_payment_intent_id;

DROP INDEX IF EXISTS idx_orders_stripe_session_id;

DROP INDEX IF EXISTS idx_orders_customer_email;

DROP INDEX IF EXISTS idx_orders_created_at;

DROP INDEX IF EXISTS idx_orders_status;

-- Drop tables
DROP TABLE IF EXISTS payments;

DROP TABLE IF EXISTS orders;

-- Drop ENUM types
DROP TYPE IF EXISTS payment_status;

DROP TYPE IF EXISTS order_status;