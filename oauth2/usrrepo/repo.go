package usrrepo

import (
	"context"

	"github.com/baozhenglab/oauth-service/oauth2/model"
	"github.com/baozhenglab/oauth-service/oauth2/storage"
)

type Storage interface {
	Find(ctx context.Context, cond map[string]interface{}) (u *storage.UserSql, err error)
	FindWithOrCond(ctx context.Context, cond map[string]interface{}, orCond map[string]interface{}) (u *storage.UserSql, err error)
	FindWithFbIdAndEmail(ctx context.Context, fbId, email string) (u *storage.UserSql, err error)
	FindWithAppleIdAndEmail(ctx context.Context, appleId, email string) (u *storage.UserSql, err error)
	FindWithAccountKit(ctx context.Context, akId, prefix, phone, email string) (u *storage.UserSql, err error)
	Create(ctx context.Context, input *storage.UserSql) (u *storage.UserSql, err error)
	Update(ctx context.Context, cond, update map[string]interface{}) error
	Updates(ctx context.Context, cond map[string]interface{}, update *model.UserUpdate) error
	Delete(ctx context.Context, uid string) error
}

type SystemManager interface {
	GetSystemSecret() string
}

type userRepository struct {
	storage Storage
	sm      SystemManager
}

func New(s Storage, sm SystemManager) *userRepository {
	return &userRepository{
		storage: s,
		sm:      sm,
	}
}
