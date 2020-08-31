package oauth2

import (
	"github.com/200lab/oauth-service/config"
	"github.com/200lab/oauth-service/oauth2/model"
	"github.com/ory/fosite"
	"time"

	"github.com/ory/fosite/compose"
)

func getStrategy(config *config.Config) compose.CommonStrategy {
	rsaKey, err := config.GetPrivateKey()

	if err != nil {
		panic(err)
	}

	return compose.CommonStrategy{
		// alternatively you could use:
		//CoreStrategy: compose.NewOAuth2HMACStrategy(config, []byte("some-super-cool-secret-that-nobody-knows"), nil),
		CoreStrategy: compose.NewOAuth2JWTStrategy(rsaKey, compose.NewOAuth2HMACStrategy(config.FC, []byte(config.SystemSecret), nil)),
		// open id connect strategy
		//OpenIDConnectTokenStrategy: compose.NewOpenIDConnectStrategy(cfg.FC, rsaKey),
	}
}

var oauth2 fosite.OAuth2Provider

func InitOAuth2Provider(config *config.Config, store interface{}) {
	strat := getStrategy(config)

	oauth2 = compose.Compose(
		config.FC,
		store,
		strat,
		nil,

		// enabled handlers
		compose.OAuth2AuthorizeExplicitFactory,
		compose.OAuth2AuthorizeImplicitFactory,
		compose.OAuth2ClientCredentialsGrantFactory,
		compose.OAuth2RefreshTokenGrantFactory,
		ResourceOwnerPasswordCredentialsFactory, // 200lab custom flow

		compose.OAuth2TokenRevocationFactory,
		compose.OAuth2TokenIntrospectionFactory,

		// be aware that open id connect factories need to be added after oauth2 factories to work properly.
		//compose.OpenIDConnectExplicitFactory,
		//compose.OpenIDConnectImplicitFactory,
		//compose.OpenIDConnectHybridFactory,
		//compose.OpenIDConnectRefreshFactory,
	)
}

func GetHasher() fosite.Hasher {
	return oauth2.(*fosite.Fosite).Hasher
}

func newSession(subject string) *model.Session {
	return &model.Session{
		DefaultSession: &fosite.DefaultSession{
			Username: subject,
			Subject:  subject,
			ExpiresAt: map[fosite.TokenType]time.Time{
				fosite.AccessToken:  time.Now().UTC().Add(time.Hour * 24 * 30),
				fosite.RefreshToken: time.Now().UTC().Add(time.Hour * 24 * 60),
			},
		},
		Extra: map[string]interface{}{},
	}
}

func newSessionForPasswordGrant(subject string, userID string) *model.Session {
	s := newSession(subject)
	s.Extra["user_id"] = userID

	return s
}
