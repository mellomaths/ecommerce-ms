package orders

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	repo "github.com/mellomaths/ecommerce-ms/internal/adapters/postgresql/sqlc"
	"github.com/mellomaths/ecommerce-ms/internal/products"
)

var (
	ErrProductNoStock = errors.New("product has not enough stock")
	ErrInvalidOrder   = errors.New("invalid order")
)

type CreateOrderParams struct {
	CustomerId int64              `json:"customer_id"`
	Items      []OrderItemsParams `json:"items"`
}

type OrderItemsParams struct {
	ProductId int64 `json:"product_id"`
	Quantity  int32 `json:"quantity"`
}

type OrderCompleted struct {
	Order             repo.Order       `json:"order"`
	Items             []repo.OrderItem `json:"items"`
	TotalPriceInCents int64            `json:"total_price_in_cents"`
}

type Service interface {
	PlaceOrder(ctx context.Context, op CreateOrderParams) (repo.Order, error)
	FindOrderById(ctx context.Context, id int64) (OrderCompleted, error)
}

type svc struct {
	repo            *repo.Queries
	db              *pgx.Conn
	productsService products.Service
}

func NewService(repo *repo.Queries, db *pgx.Conn, ps products.Service) Service {
	return &svc{repo: repo, db: db, productsService: ps}
}

func (s *svc) PlaceOrder(ctx context.Context, op CreateOrderParams) (repo.Order, error) {
	if op.CustomerId == 0 {
		return repo.Order{}, ErrInvalidOrder
	}
	if len(op.Items) == 0 {
		return repo.Order{}, ErrInvalidOrder
	}
	// transactional
	// 1. create order
	// 2. look for the product if exists
	// 3. create order items
	tx, err := s.db.Begin(ctx) // begin transaction
	if err != nil {
		return repo.Order{}, err
	}
	defer tx.Rollback(ctx) // if anything goes wrong, rollback
	qtx := s.repo.WithTx(tx)
	order, err := qtx.CreateOrder(ctx, op.CustomerId)
	if err != nil {
		return repo.Order{}, err
	}
	for _, item := range op.Items {
		product, err := s.repo.FindProductById(ctx, item.ProductId)
		if err != nil {
			return repo.Order{}, products.ErrProductNotFound
		}
		if product.Quantity < item.Quantity {
			return repo.Order{}, ErrProductNoStock
		}
		_, err = qtx.CreateOrderItem(ctx, repo.CreateOrderItemParams{
			OrderID:    order.ID,
			ProductID:  product.ID,
			Quantity:   item.Quantity,
			PriceCents: product.PriceInCents,
		})
		_, err = s.productsService.RemoveProductStock(ctx, product.ID, item.Quantity)
		if err != nil {
			return repo.Order{}, err
		}
	}
	tx.Commit(ctx)
	return order, nil
}

func (s *svc) FindOrderById(ctx context.Context, id int64) (OrderCompleted, error) {
	rows, err := s.repo.FindOrderById(ctx, id)
	if err != nil {
		return OrderCompleted{}, err
	}
	o := OrderCompleted{
		Order:             repo.Order{},
		Items:             []repo.OrderItem{},
		TotalPriceInCents: 0,
	}
	for _, r := range rows {
		o.Order = repo.Order{
			ID:         r.OrderID,
			CustomerID: r.CustomerID,
			CreatedAt:  r.CreatedAt,
		}
		i := repo.OrderItem{
			ID:         r.OrderItemID.Int64,
			OrderID:    r.OrderID,
			ProductID:  r.ProductID.Int64,
			Quantity:   r.Quantity.Int32,
			PriceCents: r.PriceCents.Int32,
		}
		o.Items = append(o.Items, i)
		o.TotalPriceInCents += int64(r.Quantity.Int32) * int64(r.PriceCents.Int32)
	}
	return o, nil
}
