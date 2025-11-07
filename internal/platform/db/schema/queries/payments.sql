-- name: CreatePayment :one
INSERT INTO payments (
        order_id,
        stripe_payment_intent_id,
        status,
        total_cents,
        currency,
        paid_at
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListPayments :many
SELECT *
FROM payments
WHERE order_id = ANY($1::uuid []);

-- name: GetOrderPayments :many
SELECT *
FROM payments
WHERE order_id = $1;