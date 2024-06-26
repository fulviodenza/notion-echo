package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/adapters/ent/user"
)

type UserRepoInterface interface {
	GetUser(ctx context.Context, id int) (*ent.User, error)
	SaveUser(ctx context.Context, id int, stateToken string) (*ent.User, error)
	GetStateTokenById(ctx context.Context, id int) (*ent.User, error)
	SaveNotionTokenByStateToken(ctx context.Context, notionToken, stateToken string) (*ent.User, error)
	GetNotionTokenByID(ctx context.Context, id int) (string, error)
	SetDefaultPage(ctx context.Context, id int, page string) error
	GetDefaultPage(ctx context.Context, id int) (string, error)
	DeleteUser(ctx context.Context, id int) error
	GetAllUsers(ctx context.Context) ([]*ent.User, error)
}

var _ UserRepoInterface = (*UserRepo)(nil)

type UserRepo struct {
	*ent.Client
}

func (ur *UserRepo) GetUser(ctx context.Context, id int) (*ent.User, error) {
	return ur.User.Get(ctx, id)
}

func (ur *UserRepo) SaveUser(ctx context.Context, id int, stateToken string) (*ent.User, error) {
	u, err := ur.User.
		Create().
		SetID(id).
		SetStateToken(stateToken).
		Save(ctx)
	if err != nil {
		isAlreadyExistsErr := strings.Contains(err.Error(), "duplicate key value violates unique constraint")
		if !isAlreadyExistsErr {
			return nil, fmt.Errorf("failed creating user: %w", err)
		}

		err := ur.User.Update().SetStateToken(stateToken).Where(user.IDEQ(id)).Exec(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed creating user: %w", err)
		}
		return u, nil
	}

	return u, nil
}

func (ur *UserRepo) GetAllUsers(ctx context.Context) ([]*ent.User, error) {
	return ur.User.Query().All(ctx)
}

func (ur *UserRepo) GetStateTokenById(ctx context.Context, id int) (*ent.User, error) {
	u, err := ur.User.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ur *UserRepo) SaveNotionTokenByStateToken(ctx context.Context, notionToken, stateToken string) (*ent.User, error) {
	err := ur.User.Update().SetNotionToken(notionToken).Where(user.StateTokenEQ(stateToken)).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (ur *UserRepo) GetNotionTokenByID(ctx context.Context, id int) (string, error) {
	u, err := ur.User.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return u.NotionToken, nil
}

func (ur *UserRepo) SetDefaultPage(ctx context.Context, id int, page string) error {
	err := ur.User.Update().SetDefaultPage(page).Where(user.IDEQ(id)).Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRepo) GetDefaultPage(ctx context.Context, id int) (string, error) {
	u, err := ur.User.Get(ctx, id)
	if err != nil {
		return "", err
	}
	return u.DefaultPage, nil
}

func (ur *UserRepo) DeleteUser(ctx context.Context, id int) error {
	_, err := ur.User.Delete().Where(user.IDEQ(id)).Exec(ctx)
	return err
}
