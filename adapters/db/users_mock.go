package db

import (
	"context"

	"github.com/notion-echo/adapters/ent"
	notionerrors "github.com/notion-echo/errors"
)

var _ UserRepoInterface = (*UserRepoMock)(nil)

type UserRepoMock struct {
	Db  map[int]*ent.User
	Err error
}

func NewUserRepoMock(db map[int]*ent.User, err error) UserRepoInterface {
	return &UserRepoMock{
		Db:  db,
		Err: err,
	}
}

func (ur *UserRepoMock) GetUser(ctx context.Context, id int) (*ent.User, error) {
	return ur.Db[id], nil
}

func (ur *UserRepoMock) SaveUser(ctx context.Context, id int, stateToken string) (*ent.User, error) {
	if ur.Err != nil {
		return nil, ur.Err
	}
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

func (ur *UserRepoMock) GetAllUsers(ctx context.Context) ([]*ent.User, error) {
	if ur.Err != nil {
		return nil, ur.Err
	}
	users := make([]*ent.User, 0)
	for _, v := range ur.Db {
		users = append(users, v)
	}
	return users, nil
}

func (ur *UserRepoMock) SaveNotionTokenByStateToken(ctx context.Context, notionToken, stateToken string) (*ent.User, error) {
	return nil, nil
}
func (ur *UserRepoMock) GetNotionTokenByID(ctx context.Context, id int) (string, error) {
	return "", nil
}
func (ur *UserRepoMock) SetDefaultPage(ctx context.Context, id int, page string) error {
	if ur.Err != nil {
		return ur.Err
	}
	return nil
}
func (ur *UserRepoMock) GetDefaultPage(ctx context.Context, id int) (string, error) {
	if ur.Err != nil {
		return "", ur.Err
	}
	if p, ok := ur.Db[id]; ok {
		return p.DefaultPage, nil
	}

	return "", notionerrors.ErrPageNotFound
}
func (ur *UserRepoMock) DeleteUser(ctx context.Context, id int) error {
	if ur.Err != nil {
		return ur.Err
	}
	ur.Db[id] = nil
	return nil
}
