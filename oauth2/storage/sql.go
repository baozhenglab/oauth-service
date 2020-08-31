package storage

import (
	"context"
	"fmt"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/oauth2"
	"github.com/200lab/oauth-service/secure"
	"github.com/jinzhu/gorm"
	"github.com/ory/fosite"
)

const (
	TbClient      = "oauth_clients"
	TbUser        = "oauth_users"
	TbAuthCode    = "oauth_authorize_codes"
	TbAccessToken = "oauth_access_tokens"
)

type DbConnectionManager interface {
	GetDB() *gorm.DB
	GetRDB() *gorm.DB
}

type sqlStore struct {
	db        DbConnectionManager
	eas       *secure.AES
	secretKey string
	Implicit  map[string]fosite.Requester // still not converted to mongo yet
}

func NewSqlStore(db DbConnectionManager, eas *secure.AES, secretKey string) *sqlStore {
	return &sqlStore{
		db:        db,
		eas:       eas,
		secretKey: secretKey,
		Implicit:  map[string]fosite.Requester{},
	}
}

type AuthorizeCodeSql struct {
	Code           string `gorm:"code"`
	ClientID       string `gorm:"client_id"`
	Active         bool   `gorm:"active"`
	Requester      *RequesterSql
	sdkcm.SQLModel `json:",inline"`
}

func (store *sqlStore) GetClient(_ context.Context, id string) (fosite.Client, error) {
	db := store.db.GetDB()

	if store.db.GetRDB() != nil {
		db = store.db.GetRDB()
	}

	db = db.New()

	var client ClientSQL

	if err := db.Table(TbClient).Where("client_id = ?", id).First(&client).Error; err != nil {
		return nil, fosite.ErrNotFound
	}

	return client.toClient(), nil
}

func (store *sqlStore) CreateAuthorizeCodeSession(_ context.Context, code string, req fosite.Requester) error {
	db := store.db.GetDB().New()

	reqSql, err := toRequesterSql(req, code)
	if err != nil {
		return err
	}

	authorizeCode := AuthorizeCodeSql{Code: code, Active: true, ClientID: reqSql.Client, Requester: reqSql}
	authorizeCode.SQLModel = *sdkcm.NewSQLModelWithStatus(1)

	if err := db.Table(TbAuthCode).Create(&authorizeCode).Error; err != nil {
		return err
	}

	return nil
}

func (store *sqlStore) GetAuthorizeCodeSession(_ context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	db := store.db.GetDB()

	if store.db.GetRDB() != nil {
		db = store.db.GetRDB()
	}

	db = db.New()

	authorizeCode := AuthorizeCodeSql{Code: code}

	if err := db.Table(TbAuthCode).Where("code = ?", code).First(&authorizeCode).Error; err != nil {
		return nil, err
	}

	req, err := authorizeCode.Requester.toRequester(store.eas, session, ClientManager(store))

	if err != nil {
		return nil, err
	}

	if !authorizeCode.Active {
		return req, fosite.ErrInvalidatedAuthorizeCode
	}

	return req, nil
}

func (store *sqlStore) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	db := store.db.GetDB().New()

	if err := db.Table(TbAuthCode).
		Where("code = ?", code).Update(map[string]interface{}{"active": 0}).Error; err != nil {
		return err
	}

	return nil
}

func (store *sqlStore) DeleteAuthorizeCodeSession(_ context.Context, code string) error {
	db := store.db.GetDB().New()
	return db.Table(TbAuthCode).Where("code = ?", code).Delete(nil).Error
}

func (store *sqlStore) CreateAccessTokenSession(ctx context.Context, signature string, req fosite.Requester) error {
	return store.createToken(ctx, signature, req, fosite.AccessToken)
}

func (store *sqlStore) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return store.getTokenSession(ctx, signature, session)
}

func (store *sqlStore) DeleteAccessTokenSession(_ context.Context, signature string) error {
	db := store.db.GetDB().New()
	return db.Table(TbAccessToken).Where("signature = ?", signature).Delete(nil).Error
}

func (store *sqlStore) CreateRefreshTokenSession(ctx context.Context, signature string, req fosite.Requester) error {
	return store.createToken(ctx, signature, req, fosite.RefreshToken)
}

func (store *sqlStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return store.getTokenSession(ctx, signature, session)
}

func (store *sqlStore) DeleteRefreshTokenSession(_ context.Context, signature string) error {
	db := store.db.GetDB().New()

	return db.Table(TbAccessToken).Where("signature = ?", signature).Delete(nil).Error
}

func (store *sqlStore) CreateImplicitAccessTokenSession(_ context.Context, code string, req fosite.Requester) error {
	store.Implicit[code] = req
	return nil
}

func (store *sqlStore) Authenticate(context context.Context, name string, secret string) (oauth2.UserCredential, error) {
	db := store.db.GetDB()

	if store.db.GetRDB() != nil {
		db = store.db.GetRDB()
	}

	db = db.New()

	var u UserSql

	if err := db.Table(TbUser).Where("username = ?", name).First(&u).Error; err != nil {
		return nil, fosite.ErrNotFound
	}

	u.UserId = fmt.Sprintf("%d", u.SQLModel.ID)

	passHash := secure.ComputeHmac256(secret, u.Salt, store.secretKey)
	if passHash != u.Password {
		return nil, fosite.ErrNotFound
	}

	u.Password = ""
	u.Salt = ""
	return u.User, nil
}

func (store *sqlStore) RevokeRefreshToken(ctx context.Context, requestID string) error {
	return store.revoke(ctx, requestID)
}

func (store *sqlStore) RevokeAccessToken(ctx context.Context, requestID string) error {
	return store.revoke(ctx, requestID)
}

func (store *sqlStore) createToken(_ context.Context, signature string, req fosite.Requester, tkt fosite.TokenType) error {
	db := store.db.GetDB().New()

	reqSql, err := toRequesterSql(req, "")
	if err != nil {
		return err
	}

	ats := AccessTokenSql{
		Signature: signature,
		Owner:     reqSql.Subject,
		Type:      string(tkt),
		ClientId:  req.GetClient().GetID(),
		ExpiredAt: req.GetSession().GetExpiresAt(fosite.AccessToken),
		RequestId: req.GetID(),
		Requester: reqSql,
	}

	ats.SQLModel = *sdkcm.NewSQLModelWithStatus(1)

	if err := db.Table(TbAccessToken).Create(&ats).Error; err != nil {
		return err
	}

	//if reqSql.Subject != "" {
	//	t := fmt.Sprintf("`type` = '%s'", string(tkt))
	//	sql := fmt.Sprintf(`delete from oauth_access_tokens
	//		where id in
	//			(select id  from
	//				(select id,row_number() over (order by id desc) as row_num
	//				from oauth_access_tokens
	//				where owner='%s' and %s) tmp
	//			where row_num > 4)`, reqSql.Subject, t)
	//
	//	db.New().Exec(sql)
	//}

	return nil
}

func (store *sqlStore) getTokenSession(_ context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	db := store.db.GetDB()

	if store.db.GetRDB() != nil {
		db = store.db.GetRDB()
	}

	db = db.New()

	var atm AccessTokenSql

	if err := db.Table(TbAccessToken).Where("signature = ?", signature).First(&atm).Error; err != nil {
		return nil, fosite.ErrNotFound
	}

	req, err := atm.Requester.toRequester(store.eas, session, store)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (store *sqlStore) revoke(ctx context.Context, requestID string) error {
	db := store.db.GetDB().New()

	if err := db.Table(TbAccessToken).Where("request_id = ?", requestID).Delete(nil).Error; err != nil {
		return err
	}

	return nil
}
