package main

import (
	"context"
	"log"
	"os"

	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/ent/migrate"

	_ "github.com/lib/pq"
)

func main() {
	client, err := ent.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed opening connection to postgres: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	if err := client.Schema.Create(ctx, migrate.WithGlobalUniqueID(true)); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}
}
