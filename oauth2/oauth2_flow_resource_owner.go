package oauth2

// This file is a custom Resource Owner flow base on fosite
// The different is we would like to return user model on authentication storage
// not just simple check true/false

import (
	"context"
	"github.com/200lab/oauth-service/oauth2/model"
	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
	foauth2 "github.com/ory/fosite/handler/oauth2"
	"github.com/pkg/errors"
	"time"
)

type ResourceOwnerPasswordCredentialsGrantStorage interface {
	Authenticate(context context.Context, name string, secret string) (UserCredential, error)
	foauth2.AccessTokenStorage
	foauth2.RefreshTokenStorage
}

type UserCredential interface {
	GetUserID() string
	GetUsername() string
	GetEmail() string
}

type ResourceOwnerPasswordCredentialsGrantHandler struct {
	// ResourceOwnerPasswordCredentialsGrantStorage is used to persist session data across requests.
	ResourceOwnerPasswordCredentialsGrantStorage ResourceOwnerPasswordCredentialsGrantStorage

	RefreshTokenStrategy     foauth2.RefreshTokenStrategy
	ScopeStrategy            fosite.ScopeStrategy
	AudienceMatchingStrategy fosite.AudienceMatchingStrategy

	*foauth2.HandleHelper
}

// HandleTokenEndpointRequest implements https://tools.ietf.org/html/rfc6749#section-4.3.2
func (c *ResourceOwnerPasswordCredentialsGrantHandler) HandleTokenEndpointRequest(ctx context.Context, request fosite.AccessRequester) error {
	// grant_type REQUIRED.
	// Value MUST be set to "password".
	if !request.GetGrantTypes().Exact("password") {
		return errors.WithStack(fosite.ErrUnknownRequest)
	}

	if !request.GetClient().GetGrantTypes().Has("password") {
		return errors.WithStack(fosite.ErrInvalidGrant.WithHint("The client is not allowed to use authorization grant \"password\"."))
	}

	client := request.GetClient()
	for _, scope := range request.GetRequestedScopes() {
		if !c.ScopeStrategy(client.GetScopes(), scope) {
			return errors.WithStack(fosite.ErrInvalidScope.WithHintf("The OAuth 2.0 Client is not allowed to request scope \"%s\".", scope))
		}
	}

	if err := c.AudienceMatchingStrategy(client.GetAudience(), request.GetRequestedAudience()); err != nil {
		return err
	}

	username := request.GetRequestForm().Get("username")
	password := request.GetRequestForm().Get("password")
	var user UserCredential
	var err error

	if username == "" || password == "" {
		return errors.WithStack(fosite.ErrInvalidRequest.WithHint("Username or password are missing from the POST body."))
	} else if user, err = c.ResourceOwnerPasswordCredentialsGrantStorage.Authenticate(ctx, username, password); errors.Cause(err) == fosite.ErrNotFound {
		return errors.WithStack(fosite.ErrRequestUnauthorized.WithHint("Unable to authenticate the provided username and password credentials.").WithDebug(err.Error()))
	} else if err != nil {
		return errors.WithStack(fosite.ErrServerError.WithDebug(err.Error()))
	}

	// Credentials must not be passed around, potentially leaking to the database!
	delete(request.GetRequestForm(), "password")

	mSession := request.GetSession().(*model.Session)
	mSession.Subject = user.GetUserID()
	mSession.SetUserID(user.GetUserID())
	mSession.SetUserEmail(user.GetEmail())
	mSession.SetUsername(user.GetUsername())
	//request.GetSession().SetExpiresAt(fosite.Token, time.Now().UTC().Add(c.AccessTokenLifespan).Round(time.Second))
	if c.RefreshTokenLifespan > -1 {
		request.GetSession().SetExpiresAt(fosite.RefreshToken, time.Now().UTC().Add(c.RefreshTokenLifespan).Round(time.Second))
	}

	return nil
}

// PopulateTokenEndpointResponse implements https://tools.ietf.org/html/rfc6749#section-4.3.3
func (c *ResourceOwnerPasswordCredentialsGrantHandler) PopulateTokenEndpointResponse(ctx context.Context, requester fosite.AccessRequester, responder fosite.AccessResponder) error {
	if !requester.GetGrantTypes().Exact("password") {
		return errors.WithStack(fosite.ErrUnknownRequest)
	}

	var refresh, refreshSignature string
	if requester.GetGrantedScopes().HasOneOf("offline", "offline_access") {
		var err error
		refresh, refreshSignature, err = c.RefreshTokenStrategy.GenerateRefreshToken(ctx, requester)
		if err != nil {
			return errors.WithStack(fosite.ErrServerError.WithDebug(err.Error()))
		} else if err := c.ResourceOwnerPasswordCredentialsGrantStorage.CreateRefreshTokenSession(ctx, refreshSignature, requester.Sanitize([]string{})); err != nil {
			return errors.WithStack(fosite.ErrServerError.WithDebug(err.Error()))
		}
	}

	if err := c.IssueAccessToken(ctx, requester, responder); err != nil {
		return err
	}

	if refresh != "" {
		responder.SetExtra("refresh_token", refresh)
	}

	return nil
}

// OAuth2ResourceOwnerPasswordCredentialsFactory creates an OAuth2 resource owner password credentials grant handler and registers
// an access token, refresh token and authorize code validator.
func ResourceOwnerPasswordCredentialsFactory(config *compose.Config, storage interface{}, strategy interface{}) interface{} {
	return &ResourceOwnerPasswordCredentialsGrantHandler{
		ResourceOwnerPasswordCredentialsGrantStorage: storage.(ResourceOwnerPasswordCredentialsGrantStorage),
		HandleHelper: &foauth2.HandleHelper{
			AccessTokenStrategy:  strategy.(foauth2.AccessTokenStrategy),
			AccessTokenStorage:   storage.(foauth2.AccessTokenStorage),
			AccessTokenLifespan:  config.GetAccessTokenLifespan(),
			RefreshTokenLifespan: config.GetRefreshTokenLifespan(),
		},
		RefreshTokenStrategy:     strategy.(foauth2.RefreshTokenStrategy),
		ScopeStrategy:            config.GetScopeStrategy(),
		AudienceMatchingStrategy: config.GetAudienceStrategy(),
	}
}
