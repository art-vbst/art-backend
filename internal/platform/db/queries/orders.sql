-- name: CreateOrder :one
INSERT INTO orders (
        stripe_session_id,
        stripe_payment_intent_id,
        customer_email,
        shipping_rate_id,
        shipping_name,
        shipping_address_line1,
        shipping_address_line2,
        shipping_city,
        shipping_postal_code,
        shipping_state,
        shipping_country,
        subtotal_cents,
        shipping_cents,
        total_cents,
        currency,
        status
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16
    )
RETURNING *;

-- name: GetOrderByID :one
SELECT *
FROM orders
WHERE id = $1;

-- name: GetOrderByStripeSessionID :one
SELECT *
FROM orders
WHERE stripe_session_id = $1;

-- name: GetOrderByStripePaymentIntentID :one
SELECT *
FROM orders
WHERE stripe_payment_intent_id = $1;

-- name: ListOrders :many
SELECT *
FROM orders
ORDER BY created_at DESC;

-- name: UpdateOrderStatus :one
UPDATE orders
SET status = $2
WHERE id = $1
RETURNING *;