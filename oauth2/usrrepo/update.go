package usrrepo

import (
	"context"
	"fmt"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/common"
	"github.com/200lab/oauth-service/oauth2/model"
	"github.com/200lab/oauth-service/secure"
)

func (ur *userRepository) UpdateUser(ctx context.Context, update *model.UserUpdate) (*model.User, error) {
	if err := update.Validate(); err != nil {
		return nil, err
	}

	if update.Email != nil && *update.Email != "" {
		oldUser, err := ur.storage.Find(ctx, map[string]interface{}{
			"email":     update.Email,
			"client_id": update.ClientId,
		})

		if err != nil && err.Error() != common.ErrDataNotFound.Error() {
			return nil, sdkcm.ErrCannotFetchData(err)
		}

		if oldUser != nil && fmt.Sprintf("%d", oldUser.ID) != update.Id {
			return nil, sdkcm.ErrCustom(nil, common.ErrEmailExisted)
		}
	}

	// oldUser := &storage.UserSql{}
	var err error

	if update.Id != "" {
		_, err = ur.storage.Find(ctx, map[string]interface{}{
			"id": update.Id,
		})
	} else {
		_, err = ur.storage.Find(ctx, map[string]interface{}{
			"username":  update.Username,
			"client_id": update.ClientId,
		})
	}

	if err != nil && err.Error() != common.ErrDataNotFound.Error() {
		return nil, sdkcm.ErrCannotFetchData(err)
	}

	//if oldUser != nil {
	//	return nil, sdkcm.ErrCustom(nil, common.ErrUsernameExisted)
	//}

	if update.Password != nil && *update.Password != "" {
		salt := secure.GenerateSalt()
		update.Salt = &salt

		password := secure.ComputeHmac256(*update.Password, salt, ur.sm.GetSystemSecret())
		update.Password = &password
	} else {
		update.PasswordConfirmation = nil
		update.Salt = nil
	}

	where := map[string]interface{}{
		"id": update.Id,
	}

	if err := ur.storage.Updates(ctx, where, update); err != nil {
		return nil, sdkcm.ErrDB(err)
	}

	u, err := ur.storage.Find(ctx, where)
	if err != nil {
		return nil, sdkcm.ErrDB(err)
	}

	return &u.User, nil
}
