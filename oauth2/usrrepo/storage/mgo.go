package storage

import (
	"context"

	"github.com/baozhenglab/oauth-service/common"
	oauthStore "github.com/baozhenglab/oauth-service/oauth2/storage"
	"github.com/baozhenglab/sdkcm"
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

type mgoStorage struct {
	mgoSession *mgo.Session
}

func NewMongo(s *mgo.Session) *mgoStorage {
	return &mgoStorage{mgoSession: s}
}

func (s *mgoStorage) Find(ctx context.Context, cond map[string]interface{}) (u *oauthStore.UserMongo, err error) {
	mgoSession := s.mgoSession.New()
	defer mgoSession.Close()

	if _, ok := cond["id"]; ok {
		cond["_id"] = bson.ObjectIdHex(cond["id"].(string))
		delete(cond, "id")
	}

	var foundUser oauthStore.UserMongo

	if err := mgoSession.DB("").C(oauthStore.UsersCollection).Find(cond).One(&foundUser); err != nil {
		if err == mgo.ErrNotFound {
			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
		}
		return nil, sdkcm.ErrDB(err)
	}
	return &foundUser, nil
}

func (s *mgoStorage) FindWithFbIdAndEmail(ctx context.Context, fbId, email string) (u *oauthStore.UserMongo, err error) {
	mgoSession := s.mgoSession.New()
	defer mgoSession.Close()

	cond := bson.M{
		"$or": []bson.M{{"fb_id": fbId}, {"email": email}},
	}

	var foundUser oauthStore.UserMongo

	if err := mgoSession.DB("").C(oauthStore.UsersCollection).Find(cond).One(&foundUser); err != nil {
		if err == mgo.ErrNotFound {
			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
		}
		return nil, err
	}
	return &foundUser, nil
}

func (s *mgoStorage) FindWithAccountKit(ctx context.Context, akId, prefix, phone, email string) (u *oauthStore.UserMongo, err error) {
	mgoSession := s.mgoSession.New()
	defer mgoSession.Close()

	or := []bson.M{{"account_kit_id": akId}}
	if email != "" {
		or = append(or, bson.M{"email": email})
	}

	if prefix+phone != "" {
		phoneCond := bson.M{
			"$and": []bson.M{{"phone_prefix": prefix}, {"phone": phone}},
		}
		or = append(or, phoneCond)
	}

	cond := bson.M{
		"$or": or,
	}

	var foundUser oauthStore.UserMongo

	if err := mgoSession.DB("").C(oauthStore.UsersCollection).Find(cond).One(&foundUser); err != nil {
		if err == mgo.ErrNotFound {
			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
		}
		return nil, err
	}
	return &foundUser, nil
}

func (s *mgoStorage) FindWithAppleIdAndEmail(ctx context.Context, appleId, email string) (u *oauthStore.UserMongo, err error) {
	mgoSession := s.mgoSession.New()
	defer mgoSession.Close()

	cond := bson.M{
		"$or": []bson.M{{"apple_id": appleId}, {"email": email}},
	}

	var foundUser oauthStore.UserMongo

	if err := mgoSession.DB("").C(oauthStore.UsersCollection).Find(cond).One(&foundUser); err != nil {
		if err == mgo.ErrNotFound {
			return nil, sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
		}
		return nil, err
	}
	return &foundUser, nil
}

func (s *mgoStorage) Create(ctx context.Context, input *oauthStore.UserMongo) (u *oauthStore.UserMongo, err error) {
	mgoSession := s.mgoSession.New()
	defer mgoSession.Close()

	input.PrepareForInsert()

	if err := mgoSession.DB("").C(oauthStore.UsersCollection).Insert(input); err != nil {
		return nil, err
	}

	return input, nil
}

func (s *mgoStorage) Update(ctx context.Context, cond, update map[string]interface{}) error {
	mgoSession := s.mgoSession.New()
	defer mgoSession.Close()

	if _, ok := cond["id"]; ok {
		cond["_id"] = bson.ObjectIdHex(cond["id"].(string))
		delete(cond, "id")
	}

	return mgoSession.DB("").C(oauthStore.UsersCollection).Update(cond, bson.M{"$set": update})
}

func (s *mgoStorage) Delete(ctx context.Context, uid string) error {
	mgoSession := s.mgoSession.New()
	defer mgoSession.Close()

	userId := bson.ObjectIdHex(uid)

	if err := mgoSession.DB("").C(oauthStore.UsersCollection).Remove(bson.M{"_id": userId}); err != nil {
		return nil
	}

	_, _ = mgoSession.DB("").C(oauthStore.UsersCollection).RemoveAll(bson.M{"owner": uid})

	return nil
}
