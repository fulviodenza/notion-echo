package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/ent/migrate"
	"github.com/sirupsen/logrus"
)

func SetupAndConnectDatabase(baseConnectionString string, logger *logrus.Logger) (UserRepoInterface, error) {
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
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Info("Database unavailable, retrying...")

		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
	}

	drv := entsql.OpenDB(dialect.Postgres, db)
	client := ent.NewClient(ent.Driver(drv))

	if err := client.Schema.Create(
		context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	); err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Fatalf("failed creating schema resources: %v", err)
	}

	logger.WithFields(logrus.Fields{
		"database": baseConnectionString,
	}).Info("Database connection established")

	return &UserRepo{
		client,
	}, nil
}
