package main

import (
	"context"
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/suessflorian/pedlar/sales/internal/config"
	"github.com/suessflorian/pedlar/sales/internal/graph"
	"github.com/suessflorian/pedlar/sales/internal/graph/resolver"
	"github.com/suessflorian/pedlar/sales/internal/items"
	"github.com/suessflorian/pedlar/sales/internal/store"
	"github.com/suessflorian/pedlar/sales/pkg/keys"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Config(ctx)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	conn, err := store.Conn(ctx, cfg.DatabaseURL, "sales")
	if err != nil {
		log.Fatalf("failed to establish connection: %v", err)
	}

	holder, err := keys.NewHolder(ctx, &store.Keys{Conn: conn})
	if err != nil {
		log.Fatalf("failed to setup key holder: %v", err)
	}

	resolver := &resolver.Resolver{
		ItemsManager: items.ItemManager{Store: &store.Items{Conn: conn}},
	}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(
		graph.Config{
			Resolvers: resolver,
			Directives: graph.DirectiveRoot{
				Opaque: func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
					keys.SetCodec(obj, holder)
					res, err = next(ctx)
					if err != nil {
						return nil, err
					}
					keys.SetCodec(res, holder)
					return res, nil
				},
			},
		},
	))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", "8080")
	log.Fatal(http.ListenAndServe(":"+"8080", nil))
}
