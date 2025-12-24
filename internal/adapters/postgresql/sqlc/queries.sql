-- name: ListProducts :many
SELECT
    *
FROM
    products;

-- name: FindProductById :one
SELECT
    *
FROM
    products
WHERE
    id = $1;

-- name: CreateOrder :one
INSERT INTO orders (
  customer_id
) VALUES ($1) RETURNING *;

-- name: CreateOrderItem :one
INSERT INTO order_items (order_id, product_id, quantity, price_cents)
VALUES ($1, $2, $3, $4) RETURNING *;

-- name: FindOrderById :many
SELECT 
	o.id as order_id,
	o.customer_id as customer_id,
	o.created_at as created_at,
	oi.id as order_item_id,
	oi.product_id as product_id,
	oi.quantity as quantity,
	oi.price_cents as price_cents
FROM 
	orders as o
LEFT JOIN order_items as oi
	ON o.id = oi.order_id
WHERE o.id = $1;
