package main

import (
	"context"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/suessflorian/pedlar/sales/internal/config"
	"github.com/suessflorian/pedlar/sales/internal/db"
	"github.com/suessflorian/pedlar/sales/internal/graph"
	"github.com/suessflorian/pedlar/sales/internal/graph/resolver"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Config(ctx)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	conn, err := db.Conn(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to establish connection: %v", err)
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: &resolver.Resolver{Conn: conn}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", "8080")
	log.Fatal(http.ListenAndServe(":"+"8080", nil))
}
