package usrrepo

import (
	"context"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/common"
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
