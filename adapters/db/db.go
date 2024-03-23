package db

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/notion-echo/adapters/ent"
)

func SetupAndConnectDatabase(baseConnectionString string) (UserRepoInterface, error) {
	var db *sql.DB
	var err error

	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = sql.Open("pgx", baseConnectionString)
		if err == nil {
			err = db.Ping()
			if err == nil {
				break
			}
		}

		log.Printf("\n[SetupAndConnectDatabase]: Database unavailable, retrying...")
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatalf("failed creating schema resources: %v", err)
	}

	log.Printf("\n[SetupAndConnectDatabase]: Database connection established")
	return &UserRepo{
		client,
	}, nil
}
