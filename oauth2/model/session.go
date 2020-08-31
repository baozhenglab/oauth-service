package model

import (
	"github.com/mohae/deepcopy"
	"github.com/ory/fosite"
	"github.com/ory/fosite/token/jwt"
	"time"
)

// A session is passed from the `/auth` to the `/token` endpoint. You probably want to store data like: "Who made the request",
// "What organization does that person belong to" and so on.
// For our use case, the session will meet the requirements imposed by JWT access tokens, HMAC access tokens and OpenID Connect
// ID Tokens plus a custom field

// newSession is a helper function for creating a new session. This may look like a lot of code but since we are
// setting up multiple strategies it is a bit longer.
// Usually, you could do:
//
//  session = new(fosite.DefaultSession)
// Note that session is not an entity in our system, it just carry more details for requester
// and access token
type Session struct {
	*fosite.DefaultSession
	Extra map[string]interface{} `json:"extra"`
}

func NewSession(subject string) *Session {
	return &Session{
		DefaultSession: &fosite.DefaultSession{
			Username: subject,
			Subject:  subject,
			ExpiresAt: map[fosite.TokenType]time.Time{
				fosite.AccessToken:  time.Now().UTC().Add(time.Hour * 24 * 30),
				fosite.RefreshToken: time.Now().UTC().Add(time.Hour * 24 * 32), // two more days than access token
			},
		},
		Extra: map[string]interface{}{},
	}
}

func (s *Session) GetJWTClaims() jwt.JWTClaimsContainer {
	claims := &jwt.JWTClaims{
		Subject: s.Subject,
		//Issuer:    s.DefaultSession.Claims.Issuer,
		Extra:     s.Extra,
		ExpiresAt: s.GetExpiresAt(fosite.AccessToken),
		IssuedAt:  time.Now(),
		NotBefore: time.Now(),
	}

	if claims.Extra == nil {
		claims.Extra = map[string]interface{}{}
	}

	//claims.Extra["client_id"] = s.ClientID
	return claims
}

func (s *Session) GetJWTHeader() *jwt.Headers {
	return &jwt.Headers{
		Extra: map[string]interface{}{},
	}
}

func (s *Session) Clone() fosite.Session {
	if s == nil {
		return nil
	}

	return deepcopy.Copy(s).(fosite.Session)
}

// A custom function to inject user id to session
func (s *Session) SetUserID(userID string) {
	s.Extra["user_id"] = userID
}

func (s *Session) SetUserEmail(email string) {
	s.Extra["email"] = email
}

func (s *Session) SetUsername(username string) {
	s.Extra["username"] = username
}

func (s *Session) GetUserID() string {
	uid, ok := s.Extra["user_id"]
	if ok {
		return uid.(string)
	}

	return ""
}

func (s *Session) GetEmail() string {
	email, ok := s.Extra["email"]
	if ok {
		return email.(string)
	}

	return ""
}
