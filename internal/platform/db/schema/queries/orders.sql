-- name: CreateOrder :one
WITH new_order AS (
    INSERT INTO orders (status, stripe_session_id)
    VALUES ($1, $2)
    RETURNING *
),
new_payment_requirement AS (
    INSERT INTO payment_requirements (
            order_id,
            subtotal_cents,
            shipping_cents,
            total_cents,
            currency
        )
    VALUES (
            (
                SELECT id
                FROM new_order
            ),
            $3,
            $4,
            $5,
            $6
        )
    RETURNING *
),
new_shipping_details AS (
    INSERT INTO shipping_details (
            order_id,
            email,
            name,
            line1,
            line2,
            city,
            state,
            postal,
            country
        )
    VALUES (
            (
                SELECT id
                FROM new_order
            ),
            $7,
            $8,
            $9,
            $10,
            $11,
            $12,
            $13,
            $14
        )
    RETURNING *
)
SELECT new_order.id as order_id,
    new_order.stripe_session_id,
    new_order.status,
    new_order.created_at,
    new_payment_requirement.id as payment_requirement_id,
    new_payment_requirement.subtotal_cents,
    new_payment_requirement.shipping_cents,
    new_payment_requirement.total_cents,
    new_payment_requirement.currency,
    new_shipping_details.id as shipping_details_id,
    new_shipping_details.email,
    new_shipping_details.name,
    new_shipping_details.line1,
    new_shipping_details.line2,
    new_shipping_details.city,
    new_shipping_details.state,
    new_shipping_details.postal,
    new_shipping_details.country
FROM new_order,
    new_payment_requirement,
    new_shipping_details;

-- name: ListOrders :many
SELECT *
FROM orders;

-- name: ListShippingDetails :many
SELECT *
FROM shipping_details;

-- name: ListPaymentRequirements :many
SELECT *
FROM payment_requirements;

-- name: GetOrder :one
SELECT *
FROM orders
WHERE id = $1;

-- name: GetOrderShippingDetail :one
SELECT *
FROM shipping_details
WHERE order_id = $1;

-- name: GetOrderPaymentRequirement :one
SELECT *
FROM payment_requirements
WHERE order_id = $1;

-- name: UpdateOrderAndShipping :one
WITH updated_order AS (
    UPDATE orders
    SET status = $2,
        stripe_session_id = $3
    WHERE orders.id = $1
    RETURNING *
),
updated_shipping_details AS (
    UPDATE shipping_details
    SET email = $4,
        name = $5,
        line1 = $6,
        line2 = $7,
        city = $8,
        state = $9,
        postal = $10,
        country = $11
    WHERE order_id = $1
    RETURNING *
)
SELECT updated_order.id as order_id,
    updated_order.stripe_session_id,
    updated_order.status,
    updated_order.created_at,
    updated_shipping_details.id as shipping_details_id,
    updated_shipping_details.email,
    updated_shipping_details.name,
    updated_shipping_details.line1,
    updated_shipping_details.line2,
    updated_shipping_details.city,
    updated_shipping_details.state,
    updated_shipping_details.postal,
    updated_shipping_details.country
FROM updated_order
    CROSS JOIN updated_shipping_details;

-- name: DeleteOrder :exec
DELETE FROM orders
WHERE id = $1;