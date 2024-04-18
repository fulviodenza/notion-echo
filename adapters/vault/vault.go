package vault

import (
	"encoding/base64"
	"os"

	vault "github.com/hashicorp/vault/api"
	"github.com/notion-echo/utils"
	"github.com/sirupsen/logrus"
)

type VaultInterface interface {
	GetKey(path string) ([]byte, error)
	Logical() *vault.Logical
}

var _ VaultInterface = (*Vault)(nil)

type Vault struct {
	*vault.Client
}

func SetupVault(addr, token string, logger *logrus.Logger) VaultInterface {
	config := vault.DefaultConfig()
	config.Address = addr

	client, err := vault.NewClient(config)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("unable to initialize Vault client")
	}

	client.SetToken(token)

	return &Vault{client}
}

func (v *Vault) Logical() *vault.Logical {
	return v.Client.Logical()
}

func (v *Vault) GetKey(path string) ([]byte, error) {
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
