package usrrepo

import (
	"context"
	"fmt"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/common"
	"github.com/200lab/oauth-service/oauth2/model"
	"github.com/200lab/oauth-service/secure"
)

func (ur *userRepository) SetUsernamePassword(ctx context.Context, user *model.CredentialAndPassword) error {
	oldUser, err := ur.storage.Find(context.Background(), map[string]interface{}{
		"username":  user.Username,
		"client_id": user.ClientId,
	})

	if err != nil && err.Error() != common.ErrDataNotFound.Error() {
		return sdkcm.ErrCannotFetchData(err)
	}

	if oldUser != nil && fmt.Sprintf("%d", oldUser.ID) != user.Id {
		return sdkcm.ErrCustom(nil, common.ErrUsernameExisted)
	}

	if user.Password != nil {
		salt := secure.GenerateSalt()
		hashPass := secure.ComputeHmac256(*user.Password, salt, ur.sm.GetSystemSecret())
		user.Salt = &salt
		user.Password = &hashPass
	}

	err = ur.storage.Update(ctx, map[string]interface{}{"id": user.Id}, user.Map())

	if err != nil {
		return sdkcm.ErrDB(err)
	}

	return nil
}
