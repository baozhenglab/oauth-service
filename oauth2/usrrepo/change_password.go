package usrrepo

import (
	"context"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/common"
	"github.com/200lab/oauth-service/secure"
)

func (ur *userRepository) ChangePassword(ctx context.Context, clientId, uid, oldPass, newPass string) error {
	oldUser, err := ur.storage.Find(context.Background(), map[string]interface{}{
		"id":        uid,
		"client_id": clientId,
	})

	if err != nil {
		return sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
	}

	// If user is external, password is empty, don't need to check
	if oldUser.Password != "" && oldUser.Password != secure.ComputeHmac256(oldPass, oldUser.Salt, ur.sm.GetSystemSecret()) {
		return sdkcm.ErrCustom(nil, common.ErrOldPassNotCorrect)
	}

	newSalt := secure.GenerateSalt()
	newPassHash := secure.ComputeHmac256(newPass, newSalt, ur.sm.GetSystemSecret())

	if err := ur.storage.Update(
		context.Background(),
		map[string]interface{}{"id": uid},
		map[string]interface{}{"salt": newSalt, "password": newPassHash},
	); err != nil {
		return sdkcm.ErrDB(err)
	}

	return nil
}
