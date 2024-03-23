package db

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/ent/user"
)

type UserRepoInterface interface {
	SaveUser(ctx context.Context, id int, stateToken string) (*ent.User, error)
	GetStateTokenById(ctx context.Context, id int) (*ent.User, error)
	SaveNotionTokenByStateToken(ctx context.Context, notionToken, stateToken string) (*ent.User, error)
	GetNotionTokenByID(ctx context.Context, id int) (*ent.User, error)
}

var _ UserRepoInterface = (*UserRepo)(nil)

type UserRepo struct {
	*ent.Client
}

func (ur *UserRepo) SaveUser(ctx context.Context, id int, stateToken string) (*ent.User, error) {
	u, err := ur.Client.User.
		Create().
		SetID(id).
		SetStateToken(stateToken).
		Save(ctx)
	isAlreadyExistsErr := strings.Contains(err.Error(), "duplicate key value violates unique constraint")
	if err != nil && !isAlreadyExistsErr {
		return nil, fmt.Errorf("failed creating user: %w", err)
	}

	if isAlreadyExistsErr {
		err := ur.Client.User.Update().SetStateToken(stateToken).Where(user.IDEQ(id)).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed creating user: %w", err)
		}
	}
	log.Println("user was created: ", u)
	return u, nil
}

func (ur *UserRepo) GetStateTokenById(ctx context.Context, id int) (*ent.User, error) {
	u, err := ur.Client.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ur *UserRepo) SaveNotionTokenByStateToken(ctx context.Context, notionToken, stateToken string) (*ent.User, error) {
	err := ur.Client.User.Update().SetNotionToken(notionToken).Where(user.StateTokenEQ(stateToken)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (ur *UserRepo) GetNotionTokenByID(ctx context.Context, id int) (*ent.User, error) {
	u, err := ur.Client.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return u, nil
}
