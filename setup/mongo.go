package setup

import (
	"context"
	"strings"

	"github.com/baozhenglab/oauth-service/config"
	"github.com/baozhenglab/oauth-service/oauth2/model"
	"github.com/baozhenglab/oauth-service/oauth2/storage"
	"github.com/baozhenglab/oauth-service/secure"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/ory/fosite"
	"github.com/pkg/errors"
)

// init mongo is a init script implement Initializer

type initMongo struct {
	cfg InitConfig
	s   *mgo.Session
	h   fosite.Hasher
}

func NewMongo(cfg *config.Config, dbSession *mgo.Session, h fosite.Hasher) *initMongo {
	im := &initMongo{cfg, dbSession, h}
	return im
}

func (init *initMongo) LoadConfig(initCfg InitConfig) error {
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

func (init *initMongo) CanRunInitScript() bool {
	db := init.s.New()
	defer db.Close()
	var mgoModel storage.MgoModel

	err := db.DB("").C(storage.UsersCollection).Find(bson.M{
		"username": init.cfg.GetRootUsername(),
	}).One(&mgoModel)

	if err == nil || err != mgo.ErrNotFound {
		init.cfg.SetInitRootOAuthId(mgoModel.PK.Hex())
	}

	return err != nil && err == mgo.ErrNotFound
}

func (init *initMongo) Run() error {
	db := init.s.New()
	defer db.Close()

	// Create indexes for all collections
	indexes := []struct {
		ColName   string
		IndexKeys []string
	}{
		{ColName: storage.ClientsCollection, IndexKeys: []string{"id", "secret", "owner_id"}},
		{ColName: storage.AuthCodesCollection, IndexKeys: []string{"code", "client_id"}},
		{ColName: storage.UsersCollection, IndexKeys: []string{"username", "fb_id", "account_kit_id", "apple_id", "email", "phone", "phone_prefix"}},
		{ColName: storage.AccessTokensCollection, IndexKeys: []string{"signature", "request_id", "client_id", "owner", "expired_at"}},
	}

	for _, idx := range indexes {
		for _, idxKey := range idx.IndexKeys {
			if err := db.DB("").C(idx.ColName).EnsureIndexKey(idxKey); err != nil {
				return errors.WithStack(err)
			}
		}
	}

	mgoModel := storage.MgoModel{}
	mgoModel.PrepareForInsert()

	// Insert root client
	secretData, _ := init.h.Hash(context.Background(), []byte(init.cfg.GetInitClientSecret()))
	rootClient := storage.ClientMongo{
		ID:            init.cfg.GetInitClientID(),
		Name:          "root",
		Secret:        string(secretData),
		RedirectURIs:  []string{"http://localhost:3846/callback"}, // actually we don't need it
		ResponseTypes: []string{"code", "token"},
		GrantTypes:    []string{"implicit", "refresh_token", "authorization_code", "password", "client_credentials"},
		Scope:         "root offline",
		OwnerID:       mgoModel.PK.Hex(),
	}
	rootClient.PrepareForInsert()

	if err := db.DB("").C(storage.ClientsCollection).Insert(&rootClient); err != nil {
		return errors.WithStack(err)
	}

	// Insert root user
	salt := secure.GenerateSalt()

	runame := init.cfg.GetRootUsername()
	email := "core@200lab.io"
	rootUser := storage.UserMongo{
		User: model.User{
			//ID:          common.NewUID(int(mgoModel.PK.Counter()), 1, 1).String(),
			Username:    &runame,
			Password:    secure.ComputeHmac256(init.cfg.GetRootPassword(), salt, init.cfg.GetSystemSecret()),
			Salt:        salt,
			AccountType: model.AccTypeInternal,
			Email:       &email,
			ClientId:    init.cfg.GetInitClientID(),
		},
		MgoModel: mgoModel,
	}

	if err := db.DB("").C(storage.UsersCollection).Insert(&rootUser); err != nil {
		return errors.WithStack(err)
	}

	init.cfg.SetInitRootOAuthId(rootClient.MgoModel.PK.Hex())

	return nil
}
