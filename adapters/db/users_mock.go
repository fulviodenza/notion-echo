package db

import (
	"context"

	"github.com/notion-echo/adapters/ent"
	"github.com/notion-echo/errors"
)

var _ UserRepoInterface = (*UserRepoMock)(nil)

type UserRepoMock struct {
	Db      map[int]*ent.User
	errored bool
}

func NewUserRepoMock(db map[int]*ent.User) UserRepoInterface {
	return &UserRepoMock{
		Db:      db,
		errored: false,
	}
}

func (ur *UserRepoMock) SaveUser(ctx context.Context, id int, stateToken string) (*ent.User, error) {
	newUser := &ent.User{
		ID:         id,
		StateToken: stateToken,
	}
	ur.Db[id] = newUser
	return newUser, nil
}

func (ur *UserRepoMock) GetStateTokenById(ctx context.Context, id int) (*ent.User, error) {
	u, ok := ur.Db[id]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (ur *UserRepoMock) SaveNotionTokenByStateToken(ctx context.Context, notionToken, stateToken string) (*ent.User, error) {
	return nil, nil
}
func (ur *UserRepoMock) GetNotionTokenByID(ctx context.Context, id int) (string, error) {
	return "", nil
}
func (ur *UserRepoMock) SetDefaultPage(ctx context.Context, id int, page string) error {
	return nil
}
func (ur *UserRepoMock) GetDefaultPage(ctx context.Context, id int) (string, error) {
	if ur.errored {
		return "", errors.ErrPageNotFound
	}
	if p, ok := ur.Db[id]; ok {
		return p.DefaultPage, nil
	}

	return "", errors.ErrPageNotFound
}
func (ur *UserRepoMock) DeleteUser(ctx context.Context, id int) error { return nil }
