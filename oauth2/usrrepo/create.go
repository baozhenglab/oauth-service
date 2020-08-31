package usrrepo

import (
	"context"
	"fmt"

	"github.com/baozhenglab/oauth-service/common"
	"github.com/baozhenglab/oauth-service/oauth2/model"
	"github.com/baozhenglab/oauth-service/oauth2/storage"
	"github.com/baozhenglab/oauth-service/secure"
	"github.com/baozhenglab/sdkcm"
)

func (ur *userRepository) Create(ctx context.Context, user *model.User) (u *model.User, err error) {
	condMap := map[string]interface{}{}

	if user.Email != nil && *user.Email != "" {
		condMap["email"] = user.Email
	}

	if user.Username != nil && *user.Username != "" {
		condMap["username"] = user.Username
	}

	if user.Phone != nil && *user.Phone != "" {
		condMap["phone"] = user.Phone
	}

	//if user.Email != nil && *user.Email != "" {
	//	oldUser, err := ur.storage.Find(ctx, map[string]interface{}{
	//		"email":     user.Email,
	//		"client_id": user.ClientId,
	//	})
	//
	//	if err != nil && err.Error() != common.ErrDataNotFound.Error() {
	//		return nil, sdkcm.ErrCannotFetchData(err)
	//	}
	//
	//	if oldUser != nil {
	//		return nil, sdkcm.ErrCustom(nil, common.ErrEmailExisted)
	//	}
	//}
	//
	//oldUser, err := ur.storage.Find(ctx, map[string]interface{}{
	//	"username":  user.Username,
	//	"client_id": user.ClientId,
	//})

	//if err != nil && err.Error() != common.ErrDataNotFound.Error() {
	//	return nil, sdkcm.ErrCannotFetchData(err)
	//}

	oldUser, err := ur.storage.FindWithOrCond(
		ctx,
		map[string]interface{}{"client_id": user.ClientId},
		condMap,
	)

	if oldUser != nil {
		return nil, sdkcm.ErrCustom(nil, common.ErrUserExisted)
	}

	// prepare data for inserting
	user.FBId = nil // dont allow set fb id here

	if user.Password != "" {
		user.Salt = secure.GenerateSalt()
		user.Password = secure.ComputeHmac256(user.Password, user.Salt, ur.sm.GetSystemSecret())
	}

	userMgo := &storage.UserSql{User: *user}
	newUser, err := ur.storage.Create(ctx, userMgo)

	if err != nil {
		return nil, sdkcm.ErrDB(err)
	}

	newUser.UserId = fmt.Sprintf("%d", newUser.ID)
	newUser.IsNew = true

	return &newUser.User, nil
}
