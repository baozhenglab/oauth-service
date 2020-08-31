package config

import (
	"crypto/rsa"
	"crypto/x509"
	"flag"

	"github.com/baozhenglab/oauth-service/secure"
	"github.com/ory/fosite/compose"
)

const (
	StorageTypeMem      = "mem"
	StorageTypeMongo    = "mongo"
	StorageTypePostgres = "postgres"
	StorageTypeMySQL    = "mysql"
)

type Config struct {
	// 32 bytes string system secret
	SystemSecret string
	// Storage Types: mem/mongo/postgres
	StorageType string
	// AES for crypto
	aes *secure.AES
	// Private Key (base64 encoded from AES Cipher)
	privateKey string
	// Fosite config
	FC *compose.Config

	// For initialization
	initRootUsername string
	initRootPassword string
	initClientID     string
	initClientSecret string
	// After create
	initRootOAuthId string
}

func SystemConfig() *Config {
	cf := &Config{
		StorageType: StorageTypeMySQL,
		privateKey:  `1jtPrI4HqpQzut00vvzcvdDteYGgcX1qhOqbl01KCt2iCz6ZkBGpBrlrquk1eFmtyZ3yQPtPMR6-Nmto5OPXiefWfWAdfpu0YW1DjuUCoMBzw3Mr4Ts_-wYV8ULnkWt1SW-IB-AD6bycEzivM7tz2f_rgPcOwzMAMaZqbX75aci5RgG0mMmg2yIwPR1iNara8uxebd4TNqzCXmkaXO-knB9RMVCXNb3bXZn3FVaEWxArtbQcpVfxxyFU807nS3Qe8b8_A-0JFYwUeXLwsWihtARtThltMffjtgMfQyUeKsxGSduwWfnUOV0C-hTKWuCas4BdMAmCBr8ZUrQfeGDYNdeXCX8lgh4SsOaa3DxZIr4VQD7Q_PHutvQ0II8nMODIhj1i2TMgc3XkvncTvCODNaK7gal_ljwiXyUIuXTvre9ATcQWS97YrgaDaC5ho8zoSOxtJWxy3fUmdPudT9uhhtvpXC7s6jtytqqXx03-IvYgiHUDL40d4YXXjGGa5cQuUfmNgs8YvHJQW8JjVPIxhAOAgaom2amz5UE-byhEEZQHfLhKhxooaaMEN2IuHor85Xo8Tamr4TAdGnMqM3MvGjX6nVgreT-zxNpVSnJ0k4FwBmB--u1EEH_RswZKiDFl73ScrzZKog9DydcNZUUnf73eQKjz8B7RtWXuWdJneRz_QlxnmBCy8v-gEWhPcLNm0wm-0332jAkZTm-kbMVI6Ww0hcdy-aRlyHCO8a07UC39ExxYD-ydl9qU18GRNBYpuq7_ri4Xq4hG_PklNeh7kNpdG1WimNqsy_J5l2zgPatpodHuUJm_Y70f-1uiMAtQZ8FQPzCsScrI4qnzJw==`,
		FC:          new(compose.Config),
	}

	flag.StringVar(&cf.SystemSecret, "secret", "mrFPTI7EYOzt8CbcQVcUo2rIoLg97HI2", "oauth system secret key")
	flag.StringVar(&cf.initRootUsername, "init-root-username", "admin", "init root username for client oauth")
	flag.StringVar(&cf.initRootPassword, "init-root-password", "Admin@2019", "init root password for client oauth")
	flag.StringVar(&cf.initClientID, "init-client-id", "200lab", "init client id for oauth")
	flag.StringVar(&cf.initClientSecret, "init-client-secret", "secret-cannot-tell", "init client secret for oauth")

	return cf
}

func (c *Config) GetAES() *secure.AES {
	if c.aes == nil {
		c.aes = secure.NewEAS([]byte(c.SystemSecret))
	}
	return c.aes
}

func (c *Config) GetPrivateKey() (key *rsa.PrivateKey, err error) {
	pk, err := c.GetAES().Decrypt(c.privateKey)

	if err != nil {
		return nil, err
	}

	return x509.ParsePKCS1PrivateKey(pk)
}

// Implement InitConfig
func (c *Config) GetSystemSecret() string {
	return c.SystemSecret
}

func (c *Config) GetRootUsername() string {
	return c.initRootUsername
}

func (c *Config) GetRootPassword() string {
	return c.initRootPassword
}

func (c *Config) GetInitClientID() string {
	return c.initClientID
}

func (c *Config) GetInitClientSecret() string {
	return c.initClientSecret
}

// After create
func (c *Config) GetInitRootOAuthId() string {
	return c.initRootOAuthId
}

func (c *Config) SetInitRootOAuthId(id string) {
	c.initRootOAuthId = id
}
