package vault

import (
	"encoding/base64"
	"log"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/notion-echo/utils"
)

type VaultInterface interface {
	GetKey(path string) ([]byte, error)
}

var _ VaultInterface = (*Vault)(nil)

type Vault struct {
	*vault.Client
}

func SetupVault(addr, token string) Vault {
	config := vault.DefaultConfig()
	config.Address = addr

	client, err := vault.NewClient(config)
	if err != nil {
		log.Fatalf("Unable to initialize Vault client: %v", err)
	}

	client.SetToken(token)

	return Vault{client}
}

func (v Vault) GetKey(path string) ([]byte, error) {
	s, err := v.Logical().Read(path)
	if err != nil {
		return nil, err
	}
	encKey := s.Data[os.Getenv(utils.VAULT_SECRET_KEY)].(string)
	decodedKey, err := base64.StdEncoding.DecodeString(encKey)
	if err != nil {
		return nil, err
	}
	return decodedKey, nil
}
