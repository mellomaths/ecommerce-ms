package products

import (
	"context"
	"errors"

	repo "github.com/mellomaths/ecommerce-ms/internal/adapters/postgresql/sqlc"
)

var (
	ErrProductNotFound = errors.New("product not found")
)

type CreateProductParams struct {
	Name         string `json:"name"`
	PriceInCents int32  `json:"price_in_cents"`
	Quantity     int32  `json:"quantity"`
}

type Service interface {
	ListProducts(ctx context.Context) ([]repo.Product, error)
	FindProductById(ctx context.Context, id int64) (repo.Product, error)
	CreateProduct(ctx context.Context, pp CreateProductParams) (repo.Product, error)
	AddProductStock(ctx context.Context, id int64, quantity int32) (repo.Product, error)
	RemoveProductStock(ctx context.Context, id int64, quantity int32) (repo.Product, error)
}

type svc struct {
	repo repo.Querier
}

func NewService(repo repo.Querier) Service {
	return &svc{repo: repo}
}

func (s *svc) ListProducts(ctx context.Context) ([]repo.Product, error) {
	products, err := s.repo.ListProducts(ctx)
	if products == nil {
		return []repo.Product{}, err
	}
	return products, err
}

func (s *svc) FindProductById(ctx context.Context, id int64) (repo.Product, error) {
	return s.repo.FindProductById(ctx, id)
}

func (s *svc) CreateProduct(ctx context.Context, pp CreateProductParams) (repo.Product, error) {
	product, err := s.repo.CreateProduct(ctx, repo.CreateProductParams{
		Name:         pp.Name,
		PriceInCents: pp.PriceInCents,
		Quantity:     pp.Quantity,
	})
	if err != nil {
		return repo.Product{}, err
	}
	return product, nil
}

func (s *svc) AddProductStock(ctx context.Context, id int64, quantity int32) (repo.Product, error) {
	p, err := s.repo.FindProductById(ctx, id)
	if err != nil {
		return repo.Product{}, ErrProductNotFound
	}
	p.Quantity += int32(quantity)
	s.repo.UpdateProduct(ctx, repo.UpdateProductParams{
		ID:           p.ID,
		Name:         p.Name,
		PriceInCents: p.PriceInCents,
		Quantity:     p.Quantity,
	})
	return p, nil
}

func (s *svc) RemoveProductStock(ctx context.Context, id int64, quantity int32) (repo.Product, error) {
	p, err := s.repo.FindProductById(ctx, id)
	if err != nil {
		return repo.Product{}, ErrProductNotFound
	}
	p.Quantity -= int32(quantity)
	s.repo.UpdateProduct(ctx, repo.UpdateProductParams{
		ID:           p.ID,
		Name:         p.Name,
		PriceInCents: p.PriceInCents,
		Quantity:     p.Quantity,
	})
	return p, nil
}
