package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/kelseyhightower/envconfig"
)

type ApplicationConfig struct {
	AccountClientUrl string `envconfig:"ACCOUNT_CLIENT_URL"`
	CatalogClientUrl string `envconfig:"CATALOG_CLIENT_URL"`
	OrderClientUrl   string `envconfig:"ORDER_CLIENT_URL"`
}

func main() {
	var appConfig ApplicationConfig
	err := envconfig.Process("", &appConfig)
	if err != nil {
		log.Fatalf("Failed to process environment variables: %v", err)
	}

	server, err := NewGraphQLServer(appConfig.AccountClientUrl, appConfig.CatalogClientUrl, appConfig.OrderClientUrl)
	if err != nil {
		log.Fatalf("Failed to create GraphQL server: %v", err)
	}

	http.Handle("/graphql", handler.New(server.ToExecutableSchema()))
	http.Handle("/playground", playground.Handler("GraphQL Playground", "/graphql"))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
