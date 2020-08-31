package storage

import (
	"context"
	"fmt"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/common"
	"github.com/200lab/oauth-service/oauth2/model"
	oauthStore "github.com/200lab/oauth-service/oauth2/storage"
	"github.com/jinzhu/gorm"
)

type DbConnectionManager interface {
	GetDB() *gorm.DB
	GetRDB() *gorm.DB
}

type sqlStorage struct {
	db DbConnectionManager
}

func NewSQL(db DbConnectionManager) *sqlStorage {
	return &sqlStorage{db}
}

func (s *sqlStorage) Find(ctx context.Context, cond map[string]interface{}) (u *oauthStore.UserSql, err error) {
	db := s.db.GetDB()

	if s.db.GetRDB() != nil {
		db = s.db.GetRDB()
	}

	db = db.New()

	var foundUser oauthStore.UserSql

	if err := db.Table(oauthStore.TbUser).Where(cond).First(&foundUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
		}
		return nil, sdkcm.ErrDB(err)
	}

	return &foundUser, nil
}

func (s *sqlStorage) FindWithOrCond(
	ctx context.Context,
	cond map[string]interface{},
	orCond map[string]interface{},
) (u *oauthStore.UserSql, err error) {
	db := s.db.GetDB()

	if s.db.GetRDB() != nil {
		db = s.db.GetRDB()
	}

	db = db.New().Table(oauthStore.TbUser).Where(cond)

	if len(orCond) == 1 {
		db = db.Where(orCond)
	} else {
		i := 0
		for k, v := range orCond {
			if i == 0 {
				db = db.Where(fmt.Sprintf("(%s = ?", k), v)
			} else if i == len(orCond)-1 {
				db = db.Or(fmt.Sprintf("%s = ?)", k), v)
			} else {
				db = db.Or(fmt.Sprintf("%s = ?", k), v)
			}
			i++
		}
	}

	var foundUser oauthStore.UserSql

	if err := db.First(&foundUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
		}
		return nil, sdkcm.ErrDB(err)
	}

	return &foundUser, nil
}

func (s *sqlStorage) FindWithFbIdAndEmail(ctx context.Context, fbId, email string) (u *oauthStore.UserSql, err error) {
	db := s.db.GetDB()

	if s.db.GetRDB() != nil {
		db = s.db.GetRDB()
	}

	db = db.New()

	var foundUser oauthStore.UserSql

	//if email != "" {
	//	if err := db.Table(oauthStore.TbUser).
	//		Where("fb_id = ? or email = ?", fbId, email).First(&foundUser).Error; err != nil {
	//		if err == gorm.ErrRecordNotFound {
	//			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
	//		}
	//		return nil, sdkcm.ErrDB(err)
	//	}
	//}

	if err := db.Table(oauthStore.TbUser).
		Where("fb_id = ?", fbId).First(&foundUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
		}
		return nil, sdkcm.ErrDB(err)
	}

	return &foundUser, nil
}

func (s *sqlStorage) FindWithAccountKit(ctx context.Context, akId, prefix, phone, email string) (u *oauthStore.UserSql, err error) {
	db := s.db.GetDB()

	if s.db.GetRDB() != nil {
		db = s.db.GetRDB()
	}

	db = db.New().Table(oauthStore.TbUser)

	db = db.Where("account_kit_id = ?", akId)
	if email != "" && prefix+phone != "" {
		db = db.Or("email = ? OR phone = ?", email, phone)
	}

	if email != "" {
		db = db.Or("email = ?", email)
	}

	if prefix+phone != "" {
		db = db.Or("phone = ?", phone)
	}

	var foundUser oauthStore.UserSql

	if err := db.First(&foundUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sdkcm.ErrDataNotFound
		}
		return nil, err
	}

	return &foundUser, nil
}

func (s *sqlStorage) FindWithAppleIdAndEmail(ctx context.Context, appleId, email string) (u *oauthStore.UserSql, err error) {
	db := s.db.GetDB()

	if s.db.GetRDB() != nil {
		db = s.db.GetRDB()
	}

	db = db.New()

	var foundUser oauthStore.UserSql

	//if email != "" {
	//	if err := db.Table(oauthStore.TbUser).
	//		Where("apple_id = ? or email = ?", appleId, email).First(&foundUser).Error; err != nil {
	//		if err == gorm.ErrRecordNotFound {
	//			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
	//		}
	//		return nil, sdkcm.ErrDB(err)
	//	}
	//}

	if err := db.Table(oauthStore.TbUser).
		Where("apple_id = ?", appleId).First(&foundUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
		}
		return nil, sdkcm.ErrDB(err)
	}

	return &foundUser, nil
}

func (s *sqlStorage) Create(ctx context.Context, input *oauthStore.UserSql) (u *oauthStore.UserSql, err error) {
	db := s.db.GetDB().New().Table(oauthStore.TbUser)

	input.SQLModel = *sdkcm.NewSQLModelWithStatus(1)

	if err := db.Create(input).Error; err != nil {
		return nil, err
	}

	return input, nil
}

func (s *sqlStorage) Update(ctx context.Context, cond, update map[string]interface{}) error {
	db := s.db.GetDB().New().Table(oauthStore.TbUser)

	return db.Where(cond).Update(update).Error
}

func (s *sqlStorage) Updates(ctx context.Context, cond map[string]interface{}, update *model.UserUpdate) error {
	db := s.db.GetDB().New().Table(oauthStore.TbUser)

	return db.Where(cond).Update(update).Error
}

func (s *sqlStorage) Delete(ctx context.Context, uid string) error {
	db := s.db.GetDB().New().Table(oauthStore.TbUser)

	if err := db.Where("id = ?", uid).Delete(nil).Error; err != nil {
		return err
	}

	db = s.db.GetDB().New().Table(oauthStore.TbAccessToken)
	db.Where("owner = ?", uid).Delete(nil)

	return nil
}
