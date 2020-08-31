package model

import "time"

type AccessToken struct {
	// Signature of token, we don't store whole token for security
	Signature string
	Owner     string

	// Access Token type: access_token or refesh
	Type      string
	CreatedAt time.Time
}
