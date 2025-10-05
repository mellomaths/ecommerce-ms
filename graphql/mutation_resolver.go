package main

import "context"

type mutationResolver struct {
	server *GraphQLServer
}

func (r *mutationResolver) CreateAccount(ctx context.Context, input AccountInput) (*Account, error) {
	return nil, nil
}

func (r *mutationResolver) CreateProduct(ctx context.Context, input ProductInput) (*Product, error) {
	return nil, nil
}

func (r *mutationResolver) CreateOrder(ctx context.Context, input OrderInput) (*Order, error) {
	return nil, nil
}
