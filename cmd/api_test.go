package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
	repo "github.com/mellomaths/ecommerce-ms/internal/adapters/postgresql/sqlc"
	"github.com/mellomaths/ecommerce-ms/internal/orders"
	"github.com/mellomaths/ecommerce-ms/internal/products"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
)

func TestCreateAndGetProduct(t *testing.T) {
	productData := products.CreateProductParams{
		Name:         "Apple Watch",
		PriceInCents: 104900,
		Quantity:     10,
	}
	conn, err := pgxmock.NewConn()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())

	expectedRow := pgxmock.NewRows([]string{"id", "name", "price_in_cents", "quantity", "created_at"}).
		AddRow(int64(1), productData.Name, productData.PriceInCents, productData.Quantity, time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600)))
	conn.ExpectQuery("INSERT INTO products").
		WithArgs(productData.Name, productData.PriceInCents, productData.Quantity).
		WillReturnRows(expectedRow)

	productsService := products.NewService(repo.New(conn))
	productsHandler := products.NewHandler(productsService)
	r2 := chi.NewRouter()
	r2.Get("/products/{id}", productsHandler.FindProductById)
	r2.Post("/products", productsHandler.CreateProduct)
	server := httptest.NewServer(r2)
	defer server.Close()

	jsonProduct, _ := json.Marshal(productData)
	resp, err := http.Post(server.URL+"/products", "application/json", bytes.NewBuffer(jsonProduct))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdProduct repo.Product
	json.NewDecoder(resp.Body).Decode(&createdProduct)
	assert.Equal(t, productData.Name, createdProduct.Name)
	assert.NotNil(t, createdProduct.ID)
	assert.NotNil(t, createdProduct.CreatedAt)
	resp.Body.Close()

	// Create a new row set for the FindProductById query (expectedRow was consumed)
	findProductRow := pgxmock.NewRows([]string{"id", "name", "price_in_cents", "quantity", "created_at"}).
		AddRow(int64(1), productData.Name, productData.PriceInCents, productData.Quantity, time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600)))

	// Match the actual query pattern - the query is "SELECT id, name, price_in_cents, quantity, created_at FROM products WHERE id = $1"
	conn.ExpectQuery("FROM products").
		WithArgs(int64(1)).
		WillReturnRows(findProductRow)

	resp, err = http.Get(server.URL + "/products/1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedProduct repo.Product
	json.NewDecoder(resp.Body).Decode(&retrievedProduct)
	assert.Equal(t, createdProduct, retrievedProduct)
	resp.Body.Close()
}

func TestListProducts(t *testing.T) {
	conn, err := pgxmock.NewConn()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())

	conn.ExpectQuery("SELECT id, name, price_in_cents, quantity, created_at FROM products").
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "price_in_cents", "quantity", "created_at"}).
			AddRow(int64(1), "Product 1", 10000, 10, time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600))).
			AddRow(int64(2), "Product 2", 20000, 20, time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600))))
	productsService := products.NewService(repo.New(conn))
	productsHandler := products.NewHandler(productsService)
	r2 := chi.NewRouter()
	r2.Get("/products", productsHandler.ListProducts)
	server := httptest.NewServer(r2)
	defer server.Close()

	resp, err := http.Get(server.URL + "/products")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var products []repo.Product
	json.NewDecoder(resp.Body).Decode(&products)
	assert.Equal(t, 2, len(products))
	assert.Equal(t, "Product 1", products[0].Name)
	assert.Equal(t, "Product 2", products[1].Name)
	resp.Body.Close()
}

func TestCreateAndGetOrder(t *testing.T) {
	conn, err := pgxmock.NewConn()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	defer func() {
		if err := conn.ExpectationsWereMet(); err != nil {
			t.Logf("Unfulfilled mock expectations: %s", err)
		}
	}()

	conn.ExpectBegin()
	// Transaction query: CreateOrder
	conn.ExpectQuery("INSERT INTO orders").
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "customer_id", "created_at"}).
			AddRow(int64(1), int64(1), time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600))))
	// Original connection query: FindProductById (for order item validation)
	conn.ExpectQuery("FROM products").
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "price_in_cents", "quantity", "created_at"}).
			AddRow(int64(1), "Product 1", int32(10000), int32(10), time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600))))
	// Transaction query: CreateOrderItem
	conn.ExpectQuery("INSERT INTO order_items").
		WithArgs(int64(1), int64(1), int32(1), int32(10000)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price_cents"}).
			AddRow(int64(1), int64(1), int64(1), int32(1), int32(10000)))
	// Original connection query: FindProductById (called by RemoveProductStock)
	conn.ExpectQuery("FROM products").
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "price_in_cents", "quantity", "created_at"}).
			AddRow(int64(1), "Product 1", int32(10000), int32(10), time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600))))
	// Original connection query: UpdateProduct (called by RemoveProductStock) - uses RETURNING so it's QueryRow
	conn.ExpectQuery("UPDATE products").
		WithArgs(int64(1), "Product 1", int32(10000), int32(9)).
		WillReturnRows(pgxmock.NewRows([]string{"id", "name", "price_in_cents", "quantity", "created_at"}).
			AddRow(int64(1), "Product 1", int32(10000), int32(9), time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600))))
	conn.ExpectCommit()
	productsService := products.NewService(repo.New(conn))
	// Use NewServiceWithDB to pass the mock connection directly (it implements the dbConn interface)
	ordersService := orders.NewServiceWithDB(repo.New(conn), conn, productsService)
	ordersHandler := orders.NewHandler(ordersService)
	r2 := chi.NewRouter()
	r2.Post("/orders", ordersHandler.PlaceOrder)
	r2.Get("/orders/{id}", ordersHandler.FindOrderById)
	server := httptest.NewServer(r2)
	defer server.Close()
	orderParams := orders.CreateOrderParams{
		CustomerId: 1,
		Items: []orders.OrderItemsParams{
			{ProductId: 1, Quantity: 1},
		},
	}
	jsonOrder, _ := json.Marshal(orderParams)
	resp, err := http.Post(server.URL+"/orders", "application/json", bytes.NewBuffer(jsonOrder))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var createdOrder repo.Order
	json.NewDecoder(resp.Body).Decode(&createdOrder)
	assert.Equal(t, orderParams.CustomerId, createdOrder.CustomerID)
	assert.NotNil(t, createdOrder.ID)
	assert.NotNil(t, createdOrder.CreatedAt)
	resp.Body.Close()

	// FindOrderById uses a single query with LEFT JOIN - need pgtype for nullable fields
	orderItemID := pgtype.Int8{Int64: 1, Valid: true}
	productID := pgtype.Int8{Int64: 1, Valid: true}
	quantity := pgtype.Int4{Int32: 1, Valid: true}
	priceCents := pgtype.Int4{Int32: 10000, Valid: true}
	// The query is: SELECT ... FROM orders as o LEFT JOIN order_items as oi ... WHERE o.id = $1
	// Use the simplest unique pattern - "WHERE o.id" should be sufficient
	conn.ExpectQuery("WHERE o.id").
		WithArgs(int64(1)).
		WillReturnRows(pgxmock.NewRows([]string{"order_id", "customer_id", "created_at", "order_item_id", "product_id", "quantity", "price_cents"}).
			AddRow(int64(1), int64(1), time.Date(2025, 12, 24, 14, 2, 58, 452793000, time.FixedZone("", -3*3600)), orderItemID, productID, quantity, priceCents))

	resp, err = http.Get(server.URL + "/orders/1")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var retrievedOrder orders.OrderCompleted
	json.NewDecoder(resp.Body).Decode(&retrievedOrder)
	assert.Equal(t, createdOrder, retrievedOrder.Order)
	assert.Equal(t, 1, len(retrievedOrder.Items))
	assert.Equal(t, int64(1), retrievedOrder.Items[0].ProductID)
	assert.Equal(t, int32(1), retrievedOrder.Items[0].Quantity)
	assert.Equal(t, int32(10000), retrievedOrder.Items[0].PriceCents)
	resp.Body.Close()
}
