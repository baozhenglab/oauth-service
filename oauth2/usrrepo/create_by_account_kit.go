package usrrepo

import (
	"context"
	"fmt"

	"github.com/baozhenglab/oauth-service/common"
	"github.com/baozhenglab/oauth-service/oauth2/model"
	"github.com/baozhenglab/oauth-service/oauth2/storage"
	"github.com/baozhenglab/sdkcm"
)

func (ur *userRepository) CreateWithAccountKit(ctx context.Context, akId, email, prefix, phone, clientId string) (u *model.User, err error) {
	if akId == "" {
		return nil, sdkcm.ErrCustom(nil, common.ErrAKIdCannotBeEmpty)
	}

	if err := checkEmailAndPhone(prefix, phone, email); err != nil {
		return nil, err
	}

	oldUser, err := ur.storage.FindWithAccountKit(context.Background(), akId, prefix, phone, email)

	if err != nil && err.Error() != common.ErrDataNotFound.Error() {
		return nil, sdkcm.ErrCannotFetchData(err)
	}

	if oldUser != nil {
		// update to db
		if oldUser.AKId != akId ||
			(oldUser.AccountType != model.AccTypeExternal && oldUser.AccountType != model.AccTypeBoth) {
			updateData := map[string]interface{}{
				"account_kit_id": akId,
				"account_type":   model.AccTypeBoth,
			}

			if email != "" && oldUser.Email != nil && *oldUser.Email != email {
				updateData["email"] = email
			}

			if err := ur.storage.Update(context.Background(),
				map[string]interface{}{"id": oldUser.ID},
				updateData); err != nil {
				return nil, sdkcm.ErrDB(err)
			}
			oldUser.AKId = akId
			oldUser.AccountType = model.AccTypeBoth
		}
		oldUser.UserId = fmt.Sprintf("%d", oldUser.ID)
		oldUser.IsNew = false
		oldUser.HasUsernamePassword = oldUser.Username != nil && oldUser.Password != ""

		return &oldUser.User, nil
	}

	newUser := &storage.UserSql{
		User: model.User{
			AKId:        akId,
			Email:       &email,
			AccountType: model.AccTypeExternal,
			ClientId:    clientId,
			Phone:       &phone,
			PhonePrefix: &prefix,
		},
	}

	if phone == "" {
		newUser.Phone = nil
		newUser.PhonePrefix = nil
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

func checkEmailAndPhone(prefix, phone, email string) error {
	if len(phone) > 1 {
		if len(prefix) < 1 {
			return sdkcm.ErrCustom(nil, common.ErrPhonePrefixCannotBeEmpty)
		}
		return nil
	}
	if len(email) < 1 {
		return sdkcm.ErrCustom(nil, common.ErrPhoneAndEmailCannotBeEmpty)
	}
	return nil
}
