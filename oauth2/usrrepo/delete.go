package usrrepo

import (
	"context"

	"github.com/baozhenglab/oauth-service/common"
	"github.com/baozhenglab/sdkcm"
)

func (ur *userRepository) Delete(ctx context.Context, clientId, uid string) error {
	_, err := ur.storage.Find(context.Background(), map[string]interface{}{
		"id":        uid,
		"client_id": clientId,
	})

	if err != nil {
		return sdkcm.ErrWithMessage(err, common.ErrDataNotFound)
	}

	if err := ur.storage.Delete(
		context.Background(),
		uid,
	); err != nil {
		return sdkcm.ErrDB(err)
	}

	return nil
}
