package usrrepo

import (
	"context"
	"fmt"

	"github.com/baozhenglab/oauth-service/common"
	"github.com/baozhenglab/oauth-service/oauth2/model"
	"github.com/baozhenglab/oauth-service/oauth2/storage"
	"github.com/baozhenglab/sdkcm"
)

func (ur *userRepository) CreateWithGmail(ctx context.Context, email, clientId string) (u *model.User, err error) {
	if email == "" {
		return nil, sdkcm.ErrCustom(nil, common.ErrEmailCannotBeEmpty)
	}

	oldUser, err := ur.storage.Find(context.Background(), map[string]interface{}{
		"email": email,
	})

	if err != nil && err.Error() != common.ErrDataNotFound.Error() {
		return nil, sdkcm.ErrCannotFetchData(err)
	}

	if oldUser != nil {
		if oldUser.AccountType != model.AccTypeExternal && oldUser.AccountType != model.AccTypeBoth {
			if err := ur.storage.Update(context.Background(),
				map[string]interface{}{"id": oldUser.ID},
				map[string]interface{}{
					"email":        email,
					"account_type": model.AccTypeBoth,
				}); err != nil {
				return nil, sdkcm.ErrDB(err)
			}
		}
		oldUser.Email = &email
		oldUser.UserId = fmt.Sprintf("%d", oldUser.ID)

		oldUser.IsNew = false
		oldUser.HasUsernamePassword = oldUser.Username != nil && oldUser.Password != ""

		return &oldUser.User, nil
	}

	newUser := &storage.UserSql{
		User: model.User{
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
