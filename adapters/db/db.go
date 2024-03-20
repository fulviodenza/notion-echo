package db

import (
	"context"
	"database/sql"
	"log"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/notion-echo/adapters/ent"
)

func SetupAndConnectDatabase(baseConnectionString string) (UserRepoInterface, error) {
	db, err := sql.Open("pgx", baseConnectionString)
	if err != nil {
		log.Fatalf("[SetupAndConnectDatabase]: %v", err)
		return nil, err
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	log.Printf("\n[SetupAndConnectDatabase]: Database connection established")
	return &UserRepo{
		client,
	}, err
}
