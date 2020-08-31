package setup

import (
	"context"
	"fmt"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/config"
	"github.com/200lab/oauth-service/oauth2/model"
	"github.com/200lab/oauth-service/oauth2/storage"
	"github.com/200lab/oauth-service/secure"
	"github.com/jinzhu/gorm"
	"github.com/ory/fosite"
	"github.com/pkg/errors"
	"strings"
)

type DbConnectionManager interface {
	GetDB() *gorm.DB
}

type initSQL struct {
	cfg InitConfig
	db  DbConnectionManager
	h   fosite.Hasher
}

func NewSQL(cfg *config.Config, db DbConnectionManager, h fosite.Hasher) *initSQL {
	im := &initSQL{cfg, db, h}
	return im
}

func (init *initSQL) LoadConfig(initCfg InitConfig) error {
	if strings.TrimSpace(initCfg.GetRootUsername()) == "" {
		return errors.WithStack(ErrRootUsernameIsEmpty)
	}

	if strings.TrimSpace(initCfg.GetRootPassword()) == "" {
		return errors.WithStack(ErrRootPasswordIsEmpty)
	}

	if strings.TrimSpace(initCfg.GetInitClientID()) == "" {
		return errors.WithStack(ErrClientIdIsEmpty)
	}

	if strings.TrimSpace(initCfg.GetInitClientSecret()) == "" {
		return errors.WithStack(ErrClientSecretIsEmpty)
	}

	init.cfg = initCfg
	return nil
}

func (init *initSQL) CanRunInitScript() bool {
	db := init.db.GetDB().New()

	var user storage.UserSql

	err := db.Table(storage.TbUser).Where("username = ?", init.cfg.GetRootUsername()).First(&user).Error

	if err != nil && err == gorm.ErrRecordNotFound {
		init.cfg.SetInitRootOAuthId(fmt.Sprintf("%d", user.ID))
	}

	return err != nil
}

func (init *initSQL) Run() error {
	db := init.db.GetDB().New()

	// Insert root user
	salt := secure.GenerateSalt()
	runame := init.cfg.GetRootUsername()
	email := "core@200lab.io"
	rootUser := storage.UserSql{

		User: model.User{
			Username:    &runame,
			Password:    secure.ComputeHmac256(init.cfg.GetRootPassword(), salt, init.cfg.GetSystemSecret()),
			Salt:        salt,
			AccountType: model.AccTypeInternal,
			Email:       &email,
			ClientId:    init.cfg.GetInitClientID(),
		},
		SQLModel: *sdkcm.NewSQLModelWithStatus(1),
	}

	if err := db.Table(storage.TbUser).Create(&rootUser).Error; err != nil {
		return errors.WithStack(err)
	}

	// Insert root client
	var n int
	db.Table(storage.TbClient).Where("client_id = ?", init.cfg.GetInitClientID()).Count(&n)
	if n == 0 {
		secretData, _ := init.h.Hash(context.Background(), []byte(init.cfg.GetInitClientSecret()))
		rootClient := storage.ClientSQL{
			ClientID:      init.cfg.GetInitClientID(),
			Name:          "root",
			Secret:        string(secretData),
			RedirectURIs:  "http://localhost:3846/callback",
			ResponseTypes: "code, token",
			GrantTypes:    "implicit,refresh_token,authorization_code,password,client_credentials",
			Scope:         "root offline",
			OwnerID:       rootUser.ID,
			SQLModel:      *sdkcm.NewSQLModelWithStatus(1),
		}

		if err := db.Table(storage.TbClient).Create(&rootClient).Error; err != nil {
			return errors.WithStack(err)
		}
	}

	init.cfg.SetInitRootOAuthId(fmt.Sprintf("%d", rootUser.ID))
	return nil
}
