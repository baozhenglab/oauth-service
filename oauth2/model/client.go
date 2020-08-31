package model

import (
	"github.com/ory/fosite"
	"strings"
	"time"
)

// Client represents an OAuth 2.0 Client.
// We don't use fosite.DefaultClient because it lacks of many info
type Client struct {
	// ClientID  is the id for this client.
	ClientID string `json:"client_id"`

	// Name is the human-readable string name of the client to be presented to the
	// end-user during authorization.
	Name string `json:"client_name"`

	// Secret is the client's secret. The secret will be included in the create request as cleartext, and then
	// never again. The secret is stored using BCrypt so it is impossible to recover it. Tell your users
	// that they need to write the secret down as it will not be made available again.
	Secret string `json:"client_secret,omitempty"`

	// RedirectURIs is an array of allowed redirect urls for the client, for example http://mydomain/oauth/callback .
	RedirectURIs []string `json:"redirect_uris"`

	// GrantTypes is an array of grant types the client is allowed to use.
	//
	// Pattern: client_credentials|authorization_code|implicit|refresh_token
	GrantTypes []string `json:"grant_types"`

	// ResponseTypes is an array of the OAuth 2.0 response type strings that the client can
	// use at the authorization endpoint.
	//
	// Pattern: id_token|code|token
	ResponseTypes []string `json:"response_types"`

	// Scope is a string containing a space-separated list of scope values (as
	// described in Section 3.3 of OAuth 2.0 [RFC6749]) that the client
	// can use when requesting access tokens.
	//
	// Pattern: ([a-zA-Z0-9\.\*]+\s?)+
	Scope string `json:"scope"`

	// Audience is a whitelist defining the audiences this client is allowed to request tokens for. An audience limits
	// the applicability of an OAuth 2.0 Access Token to, for example, certain API endpoints. The value is a list
	// of URLs. URLs MUST NOT contain whitespaces.
	Audience []string `json:"audience"`

	// Owner is a string identifying the owner of the OAuth 2.0 Client.
	Owner string `json:"owner"`

	// PolicyURI is a URL string that points to a human-readable privacy policy document
	// that describes how the deployment organization collects, uses,
	// retains, and discloses personal data.
	PolicyURI string `json:"policy_uri"`

	// AllowedCORSOrigins are one or more URLs (scheme://host[:port]) which are allowed to make CORS requests
	// to the /oauth/token endpoint. If this array is empty, the sever's CORS origin configuration (`CORS_ALLOWED_ORIGINS`)
	// will be used instead. If this array is set, the allowed origins are appended to the server's CORS origin configuration.
	// Be aware that environment variable `CORS_ENABLED` MUST be set to `true` for this to work.
	AllowedCORSOrigins []string `json:"allowed_cors_origins"`

	// TermsOfServiceURI is a URL string that points to a human-readable terms of service
	// document for the client that describes a contractual relationship
	// between the end-user and the client that the end-user accepts when
	// authorizing the client.
	TermsOfServiceURI string `json:"tos_uri"`

	// ClientURI is an URL string of a web page providing information about the client.
	// If present, the server SHOULD display this URL to the end-user in
	// a clickable fashion.
	ClientURI string `json:"client_uri"`

	// LogoURI is an URL string that references a logo for the client.
	LogoURI string `json:"logo_uri"`

	// Contacts is a array of strings representing ways to contact people responsible
	// for this client, typically email addresses.
	Contacts []string `json:"contacts"`

	// CreatedAt returns the timestamp of the client's creation.
	CreatedAt time.Time `json:"created_at,omitempty"`

	// UpdatedAt returns the timestamp of the last update.
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

func (c *Client) GetID() string {
	return c.ClientID
}

func (c *Client) GetRedirectURIs() []string {
	return c.RedirectURIs
}

func (c *Client) GetHashedSecret() []byte {
	return []byte(c.Secret)
}

func (c *Client) GetScopes() fosite.Arguments {
	return fosite.Arguments(strings.Fields(c.Scope))
}

func (c *Client) GetAudience() fosite.Arguments {
	return fosite.Arguments(c.Audience)
}

func (c *Client) GetGrantTypes() fosite.Arguments {
	// https://openid.net/specs/openid-connect-registration-1_0.html#ClientMetadata
	//
	// JSON array containing a list of the OAuth 2.0 Grant Types that the Client is declaring
	// that it will restrict itself to using.
	// If omitted, the default is that the Client will use only the authorization_code Grant Type.
	if len(c.GrantTypes) == 0 {
		return fosite.Arguments{"authorization_code"}
	}
	return fosite.Arguments(c.GrantTypes)
}

func (c *Client) GetResponseTypes() fosite.Arguments {
	// https://openid.net/specs/openid-connect-registration-1_0.html#ClientMetadata
	//
	// <JSON array containing a list of the OAuth 2.0 response_type values that the Client is declaring
	// that it will restrict itself to using. If omitted, the default is that the Client will use
	// only the code Response Type.
	if len(c.ResponseTypes) == 0 {
		return fosite.Arguments{"code"}
	}
	return fosite.Arguments(c.ResponseTypes)
}

func (c *Client) IsPublic() bool {
	return false
}

func (c *Client) GetOwner() string {
	return c.Owner
}
