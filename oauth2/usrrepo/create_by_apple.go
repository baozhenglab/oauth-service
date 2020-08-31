package usrrepo

import (
	"context"
	"fmt"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/common"
	"github.com/200lab/oauth-service/oauth2/model"
	"github.com/200lab/oauth-service/oauth2/storage"
)

func (ur *userRepository) CreateWithApple(ctx context.Context, appleId, email, clientId string) (u *model.User, err error) {
	if appleId == "" {
		return nil, sdkcm.ErrCustom(nil, common.ErrAppleIdCannotBeEmpty)
	}

	oldUser, err := ur.storage.FindWithAppleIdAndEmail(context.Background(), appleId, email)

	if err != nil && err.Error() != common.ErrDataNotFound.Error() {
		return nil, sdkcm.ErrCannotFetchData(err)
	}

	if oldUser != nil {
		// update to db
		if oldUser.AppleId != nil && *oldUser.AppleId != appleId ||
			(oldUser.AccountType != model.AccTypeExternal && oldUser.AccountType != model.AccTypeBoth) {
			updateData := map[string]interface{}{
				"apple_id":     appleId,
				"account_type": model.AccTypeBoth,
			}

			if email != "" && oldUser.Email != nil && *oldUser.Email != email {
				updateData["email"] = email
			}

			if err := ur.storage.Update(context.Background(),
				map[string]interface{}{"id": oldUser.ID},
				updateData); err != nil {
				return nil, sdkcm.ErrDB(err)
			}
			oldUser.AppleId = &appleId
			oldUser.AccountType = model.AccTypeBoth
		}
		oldUser.UserId = fmt.Sprintf("%d", oldUser.ID)
		oldUser.IsNew = false
		oldUser.HasUsernamePassword = oldUser.Username != nil && oldUser.Password != ""

		return &oldUser.User, nil
	}

	newUser := &storage.UserSql{
		User: model.User{
			AppleId:     &appleId,
			Email:       &email,
			AccountType: model.AccTypeExternal,
			ClientId:    clientId,
		},
	}

	nu, err := ur.storage.Create(context.Background(), newUser)

	if err != nil {
		return nil, sdkcm.ErrDB(err)
	}

	nu.UserId = fmt.Sprintf("%d", newUser.ID)
	nu.IsNew = true
	nu.HasUsernamePassword = nu.Username != nil && nu.Password != ""

	return &nu.User, nil
}
