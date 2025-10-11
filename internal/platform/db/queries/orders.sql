-- name: CreateOrder :one
WITH new_order AS (
    INSERT INTO orders (
        stripe_session_id,
        status
    ) VALUES (
        $1,
        $2
    ) RETURNING *
),
new_payment_requirement AS (
    INSERT INTO payment_requirements (
        order_id,
        subtotal_cents,
        shipping_cents,
        total_cents,
        currency
    ) VALUES (
        (SELECT id FROM new_order),
        $3,
        $4,
        $5,
        $6
    ) RETURNING *
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
    ) VALUES (
        (SELECT id FROM new_order),
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14
    ) RETURNING *
)
SELECT 
    new_order.id as order_id,
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
FROM new_order, new_payment_requirement, new_shipping_details;
