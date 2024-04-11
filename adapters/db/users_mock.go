package db

import (
	"context"

	"github.com/notion-echo/adapters/ent"
)

var _ UserRepoInterface = (*UserRepoMock)(nil)

type UserRepoMock struct {
	db      map[int]*ent.User
	errored bool
}

func NewUserRepoMock(db map[int]*ent.User) UserRepoInterface {
	return &UserRepoMock{
		db:      db,
		errored: false,
	}
}

func (ur *UserRepoMock) SaveUser(ctx context.Context, id int, stateToken string) (*ent.User, error) {
	newUser := &ent.User{
		ID:         id,
		StateToken: stateToken,
	}
	ur.db[id] = newUser
	return newUser, nil
}

func (ur *UserRepoMock) GetStateTokenById(ctx context.Context, id int) (*ent.User, error) {
	u, ok := ur.db[id]
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
func (ur *UserRepoMock) SetDefaultPage(ctx context.Context, id int, page string) error { return nil }
func (ur *UserRepoMock) GetDefaultPage(ctx context.Context, id int) (string, error)    { return "", nil }
func (ur *UserRepoMock) DeleteUser(ctx context.Context, id int) error                  { return nil }
