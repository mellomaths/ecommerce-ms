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
	repo "github.com/mellomaths/ecommerce-ms/internal/adapters/postgresql/sqlc"
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
