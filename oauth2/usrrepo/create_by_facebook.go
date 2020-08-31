package usrrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/baozhenglab/oauth-service/common"
	"github.com/baozhenglab/oauth-service/oauth2/model"
	"github.com/baozhenglab/oauth-service/oauth2/storage"
	"github.com/baozhenglab/sdkcm"
)

func (ur *userRepository) CreateWithFacebook(ctx context.Context, fbId, email, clientId string) (u *model.User, err error) {
	if fbId == "" {
		return nil, sdkcm.ErrCustom(nil, common.ErrFbIdCannotBeEmpty)
	}

	email = strings.TrimSpace(email)

	oldUser, err := ur.storage.FindWithFbIdAndEmail(context.Background(), fbId, email)

	if err != nil && err.Error() != common.ErrDataNotFound.Error() {
		return nil, sdkcm.ErrCannotFetchData(err)
	}

	if oldUser != nil {
		// update to db
		if oldUser.FBId != nil && *oldUser.FBId != fbId ||
			(oldUser.AccountType != model.AccTypeExternal && oldUser.AccountType != model.AccTypeBoth) {
			updateData := map[string]interface{}{
				"fb_id":        fbId,
				"account_type": model.AccTypeBoth,
			}

			//if email != "" && oldUser.Email != nil && *oldUser.Email != email {
			//	updateData["email"] = email
			//}

			if err := ur.storage.Update(context.Background(),
				map[string]interface{}{"id": oldUser.ID},
				updateData); err != nil {
				return nil, sdkcm.ErrDB(err)
			}
			oldUser.FBId = &fbId
			oldUser.AccountType = model.AccTypeBoth
		}

		oldUser.UserId = fmt.Sprintf("%d", oldUser.ID)
		oldUser.IsNew = false
		oldUser.HasUsernamePassword = oldUser.Username != nil && oldUser.Password != ""

		return &oldUser.User, nil
	}

	newUser := &storage.UserSql{
		User: model.User{
			FBId:        &fbId,
			AccountType: model.AccTypeExternal,
			ClientId:    clientId,
		},
	}

	//if email != "" {
	//	newUser.User.Email = &email
	//}

	nu, err := ur.storage.Create(context.Background(), newUser)

	if err != nil {
		return nil, sdkcm.ErrDB(err)
	}

	nu.UserId = fmt.Sprintf("%d", newUser.ID)
	nu.IsNew = true
	nu.HasUsernamePassword = nu.Username != nil && nu.Password != ""

	return &nu.User, nil
}
