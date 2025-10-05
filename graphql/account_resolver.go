package main

import "context"

type accountResolver struct {
	server *GraphQLServer
}

func (r *accountResolver) Orders(ctx context.Context, account *Account) ([]*Order, error) {
	return nil, nil
}
