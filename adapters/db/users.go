package db

import (
	"context"
	"fmt"
	"log"

	"github.com/notion-echo/adapters/ent"
)

type UserRepoInterface interface {
	SaveUser(ctx context.Context, id int, notionToken string) (*ent.User, error)
}

var _ UserRepoInterface = (*UserRepo)(nil)

type UserRepo struct {
	*ent.Client
}

func (ur *UserRepo) SaveUser(ctx context.Context, id int, notionToken string) (*ent.User, error) {
	u, err := ur.Client.User.
		Create().
		SetID(id).
		SetNotionToken(notionToken).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed creating user: %w", err)
	}
	log.Println("user was created: ", u)
	return u, nil
}
