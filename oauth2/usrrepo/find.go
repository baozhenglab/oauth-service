package usrrepo

import (
	"context"
	"fmt"

	"github.com/baozhenglab/oauth-service/oauth2/model"
)

func (ur *userRepository) Find(ctx context.Context, filter *model.UserFilter) (u *model.User, err error) {
	sqlUser, err := ur.storage.Find(ctx, filter.Map())
	if err != nil {
		return nil, err
	}

	sqlUser.User.UserId = fmt.Sprintf("%d", sqlUser.SQLModel.ID)

	return &sqlUser.User, nil
}
