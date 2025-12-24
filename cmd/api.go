package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5"
	repo "github.com/mellomaths/ecommerce-ms/internal/adapters/postgresql/sqlc"
	"github.com/mellomaths/ecommerce-ms/internal/orders"
	"github.com/mellomaths/ecommerce-ms/internal/products"
)

type application struct {
	config config
	db     *pgx.Conn
}

func (app *application) mount() http.Handler {
	r := chi.NewRouter()

	// Middlewares
	r.Use(middleware.RequestID) // Rate Limiting
	r.Use(middleware.RealIP)    // Rate Limiting, analytics and tracing
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // Recover from crashes

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	// Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
	})

	// Product Handlers
	productsService := products.NewService(repo.New(app.db))
	productsHandler := products.NewHandler(productsService)
	r.Get("/products", productsHandler.ListProducts)
	r.Get("/products/{id}", productsHandler.FindProductById)

	// Order Handlers
	ordersService := orders.NewService(repo.New(app.db), app.db)
	ordersHandler := orders.NewHandler(ordersService)
	r.Post("/orders", ordersHandler.PlaceOrder)
	r.Get("/orders/{id}", ordersHandler.FindOrderById)

	return r
}

func (app *application) run(h http.Handler) error {
	srv := &http.Server{
		Addr:         app.config.addr,
		Handler:      h,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	log.Printf("server has started at addr %s", app.config.addr)
	return srv.ListenAndServe()
}

type config struct {
	addr string
	db   dbConfig
}

type dbConfig struct {
	dsn string
}
