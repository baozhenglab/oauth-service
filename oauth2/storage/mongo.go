package storage

import (
	"context"

	"github.com/baozhenglab/oauth-service/oauth2"
	"github.com/baozhenglab/oauth-service/secure"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/ory/fosite"
)

const (
	ClientsCollection      = "clients"
	UsersCollection        = "users"
	AuthCodesCollection    = "authorize_codes"
	AccessTokensCollection = "access_tokens"
)

type MgoConnectionManage interface {
	GetSession() *mgo.Session
}

type mongoStore struct {
	s         MgoConnectionManage
	eas       *secure.AES
	secretKey string
	Implicit  map[string]fosite.Requester // still not converted to mongo yet
}

func NewMongoStore(mgoSession MgoConnectionManage, eas *secure.AES, secretKey string) *mongoStore {
	return &mongoStore{
		s:         mgoSession,
		eas:       eas,
		secretKey: secretKey,
		Implicit:  map[string]fosite.Requester{},
	}
}

type AuthorizeCode struct {
	Code      string `bson:"code"`
	ClientID  string `bson:"client_id"`
	Active    bool   `bson:"active"`
	Requester *RequesterMongo
	MgoModel  `bson:",inline"`
}

func (store *mongoStore) GetClient(_ context.Context, id string) (fosite.Client, error) {
	s := store.s.GetSession()
	defer s.Close()

	var clientMongo ClientMongo
	if err := s.DB("").C(ClientsCollection).Find(bson.M{"id": id}).One(&clientMongo); err != nil {
		return nil, fosite.ErrNotFound
	}

	return clientMongo.toClient(), nil
}

func (store *mongoStore) CreateAuthorizeCodeSession(_ context.Context, code string, req fosite.Requester) error {
	s := store.s.GetSession()
	defer s.Close()

	reqMongo, err := toRequesterMongo(req, code)
	if err != nil {
		return err
	}

	authorizeCode := AuthorizeCode{Code: code, Active: true, ClientID: reqMongo.Client, Requester: reqMongo}
	authorizeCode.PrepareForInsert()

	if err := s.DB("").C(AuthCodesCollection).Insert(&authorizeCode); err != nil {
		return err
	}

	return nil
}

func (store *mongoStore) GetAuthorizeCodeSession(_ context.Context, code string, session fosite.Session) (fosite.Requester, error) {
	s := store.s.GetSession()
	defer s.Close()

	authorizeCode := AuthorizeCode{Code: code}

	if err := s.DB("").C(AuthCodesCollection).Find(bson.M{"code": code}).One(&authorizeCode); err != nil {
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

func (store *mongoStore) InvalidateAuthorizeCodeSession(ctx context.Context, code string) error {
	s := store.s.GetSession()
	defer s.Close()

	if err := s.DB("").C(AuthCodesCollection).Update(bson.M{"code": code}, bson.M{
		"$set": bson.M{
			"active": false,
		},
	}); err != nil {
		return err
	}

	return nil
}

func (store *mongoStore) DeleteAuthorizeCodeSession(_ context.Context, code string) error {
	s := store.s.GetSession()
	defer s.Close()

	authorizeCode := AuthorizeCode{Code: code}
	if err := s.DB("").C(AuthCodesCollection).Remove(authorizeCode); err != nil {
		return err
	}

	return nil
}

//func (store *mongoStore) CreatePKCERequestSession(_ context.Context, code string, req fosite.Requester) error {
//	store.PKCES[code] = req
//	return nil
//}
//
//func (store *mongoStore) GetPKCERequestSession(_ context.Context, code string, _ fosite.Session) (fosite.Requester, error) {
//	rel, ok := store.PKCES[code]
//	if !ok {
//		return nil, fosite.ErrNotFound
//	}
//	return rel, nil
//}
//
//func (store *mongoStore) DeletePKCERequestSession(_ context.Context, code string) error {
//	delete(store.PKCES, code)
//	return nil
//}

func (store *mongoStore) CreateAccessTokenSession(ctx context.Context, signature string, req fosite.Requester) error {
	return store.createToken(ctx, signature, req, fosite.AccessToken)
}

func (store *mongoStore) GetAccessTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return store.getTokenSession(ctx, signature, session)
}

func (store *mongoStore) DeleteAccessTokenSession(_ context.Context, signature string) error {
	s := store.s.GetSession()
	defer s.Close()

	_ = s.DB("").C(AccessTokensCollection).Remove(bson.M{"signature": signature})
	return nil
}

func (store *mongoStore) CreateRefreshTokenSession(ctx context.Context, signature string, req fosite.Requester) error {
	return store.createToken(ctx, signature, req, fosite.RefreshToken)
}

func (store *mongoStore) GetRefreshTokenSession(ctx context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	return store.getTokenSession(ctx, signature, session)
}

func (store *mongoStore) DeleteRefreshTokenSession(_ context.Context, signature string) error {
	s := store.s.GetSession()
	defer s.Close()

	if err := s.DB("").C(AccessTokensCollection).Remove(bson.M{"signature": signature}); err != nil {
		return err
	}

	return nil
}

func (store *mongoStore) CreateImplicitAccessTokenSession(_ context.Context, code string, req fosite.Requester) error {
	store.Implicit[code] = req
	return nil
}

func (store *mongoStore) Authenticate(context context.Context, name string, secret string) (oauth2.UserCredential, error) {
	s := store.s.GetSession()
	defer s.Close()

	var u UserMongo

	if err := s.DB("").C(UsersCollection).Find(bson.M{"username": name}).One(&u); err != nil {
		return nil, fosite.ErrNotFound
	}

	u.UserId = u.PK.Hex()

	passHash := secure.ComputeHmac256(secret, u.Salt, store.secretKey)
	if passHash != u.Password {
		return nil, fosite.ErrNotFound
	}

	u.Password = ""
	u.Salt = ""
	return u.User, nil
}

func (store *mongoStore) RevokeRefreshToken(ctx context.Context, requestID string) error {
	return store.revoke(ctx, requestID)
}

func (store *mongoStore) RevokeAccessToken(ctx context.Context, requestID string) error {
	return store.revoke(ctx, requestID)
}

func (store *mongoStore) createToken(_ context.Context, signature string, req fosite.Requester, tkt fosite.TokenType) error {
	s := store.s.GetSession()
	defer s.Close()

	reqMongo, err := toRequesterMongo(req, "")
	if err != nil {
		return err
	}

	atm := AccessTokenMongo{
		Signature: signature,
		Owner:     reqMongo.Subject,
		Type:      string(tkt),
		ClientID:  req.GetClient().GetID(),
		ExpiredAt: req.GetSession().GetExpiresAt(fosite.AccessToken),
		RequestID: req.GetID(),
		Requester: reqMongo,
	}
	atm.PrepareForInsert()

	if err := s.DB("").C(AccessTokensCollection).Insert(&atm); err != nil {
		return err
	}

	return nil
}

func (store *mongoStore) getTokenSession(_ context.Context, signature string, session fosite.Session) (fosite.Requester, error) {
	s := store.s.GetSession()
	defer s.Close()

	var atm AccessTokenMongo

	if err := s.DB("").C(AccessTokensCollection).Find(bson.M{"signature": signature}).One(&atm); err != nil {
		return nil, fosite.ErrNotFound
	}

	req, err := atm.Requester.toRequester(store.eas, session, store)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (store *mongoStore) revoke(ctx context.Context, requestID string) error {
	s := store.s.GetSession()
	defer s.Close()

	if err := s.DB("").C(AccessTokensCollection).Remove(bson.M{"request_id": requestID}); err != nil {
		return err
	}

	return nil
}
