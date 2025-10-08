-- name: CreatePayment :one
INSERT INTO payments (
        order_id,
        stripe_payment_intent_id,
        subtotal_cents,
        shipping_cents,
        shipping_stripe_id,
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
        $8
    )
RETURNING *;

-- name: GetPaymentByID :one
SELECT *
FROM payments
WHERE id = $1;

-- name: GetPaymentByOrderID :one
SELECT *
FROM payments
WHERE order_id = $1;

-- name: GetPaymentByStripePaymentIntentID :one
SELECT *
FROM payments
WHERE stripe_payment_intent_id = $1;

-- name: ListPayments :many
SELECT *
FROM payments
ORDER BY created_at DESC;

-- name: UpdatePaymentStatus :one
UPDATE payments
SET status = $2
WHERE id = $1
RETURNING *;