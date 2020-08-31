package storage

import (
	"context"
	"github.com/ory/fosite"
)

type ClientManager interface {
	GetClient(context context.Context, clientID string) (fosite.Client, error)
}
