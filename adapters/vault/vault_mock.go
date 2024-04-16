package vault

import (
	vault "github.com/hashicorp/vault/api"
	"github.com/notion-echo/errors"
)

var _ VaultInterface = (*Vault)(nil)

type VaultMock struct {
	keys map[string]string
	err  error
}

func NewVaultMock(keys map[string]string, err error) VaultInterface {
	return &VaultMock{
		keys: keys,
		err:  err,
	}
}

func (v *VaultMock) GetKey(path string) ([]byte, error) {
	p, ok := v.keys[path]
	if !ok {
		return nil, errors.ErrNotRegistered
	}
	return []byte(p), nil
}

func (v *VaultMock) Logical() *vault.Logical {
	return nil
}
