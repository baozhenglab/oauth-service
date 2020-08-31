package setup

import "github.com/go-errors/errors"

var (
	ErrRootUsernameIsEmpty = errors.New("init root username can not be empty")
	ErrRootPasswordIsEmpty = errors.New("init root password can not be empty")
	ErrClientIdIsEmpty     = errors.New("init client id can not be empty")
	ErrClientSecretIsEmpty = errors.New("init client secret can not be empty")
)

type InitConfig interface {
	GetSystemSecret() string
	GetRootUsername() string
	GetRootPassword() string
	GetInitClientID() string
	GetInitClientSecret() string
	SetInitRootOAuthId(id string)
}

type Initializer interface {
	LoadConfig(initCfg InitConfig) error
	CanRunInitScript() bool
	Run() error
}
