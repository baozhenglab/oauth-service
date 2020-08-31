package storage

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"github.com/200lab/go-sdk/sdkcm"
	"github.com/200lab/oauth-service/oauth2/model"
	"github.com/200lab/oauth-service/secure"
	"github.com/ory/fosite"
	"github.com/ory/go-convenience/stringsx"
	"github.com/pkg/errors"
	"net/url"
	"strings"
	"time"
)

type ClientSQL struct {
	ClientID          string       `gorm:"column:client_id"`
	Name              string       `gorm:"column:client_name"`
	Secret            string       `gorm:"column:client_secret"`
	RedirectURIs      string       `gorm:"column:redirect_uris"`
	GrantTypes        string       `gorm:"column:grant_types"`
	ResponseTypes     string       `gorm:"column:response_types"`
	Scope             string       `gorm:"column:scope"`
	Audience          string       `gorm:"column:audiences"`
	OwnerID           uint32       `gorm:"column:owner_id"`
	PolicyURI         string       `gorm:"column:policy_uri"`
	TermsOfServiceURI string       `gorm:"column:tos_uri"`
	ClientURI         string       `gorm:"column:client_uri"`
	LogoURI           *sdkcm.Image `gorm:"column:logo"`
	Contacts          string       `gorm:"column:contacts"`
	SecretExpiresAt   int          `gorm:"column:client_secret_expires_at"`
	sdkcm.SQLModel    `json:",inline"`
}

func (c *ClientSQL) toClient() *model.Client {
	clt := &model.Client{
		ClientID:          c.ClientID,
		Name:              c.Name,
		Secret:            c.Secret,
		Audience:          strings.Split(c.Audience, ","),
		RedirectURIs:      strings.Split(c.RedirectURIs, ","),
		GrantTypes:        strings.Split(c.GrantTypes, ","),
		ResponseTypes:     strings.Split(c.ResponseTypes, ","),
		Scope:             c.Scope,
		Owner:             fmt.Sprintf("%d", c.OwnerID),
		PolicyURI:         c.PolicyURI,
		TermsOfServiceURI: c.TermsOfServiceURI,
		ClientURI:         c.ClientURI,
		//LogoURI:           c.LogoURI,
		Contacts: strings.Split(c.Contacts, ","),
	}

	return clt
}

type RequesterSql struct {
	Signature         string    `json:"signature"`
	Request           string    `json:"request_id"`
	RequestedAt       time.Time `json:"requested_at"`
	Client            string    `json:"client_id"`
	Scopes            string    `json:"scope"`
	GrantedScope      string    `json:"granted_scope"`
	RequestedAudience string    `json:"requested_audience"`
	GrantedAudience   string    `json:"granted_audience"`
	Form              string    `json:"form_data"`
	Subject           string    `json:"subject"`
	Session           []byte    `json:"session_data"`
}

func (rs *RequesterSql) toRequester(_ *secure.AES, session fosite.Session, clm ClientManager) (*fosite.Request, error) {
	if session != nil {
		if err := json.Unmarshal(rs.Session, session); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	c, err := clm.GetClient(context.Background(), rs.Client)
	if err != nil {
		return nil, err
	}

	val, err := url.ParseQuery(rs.Form)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r := &fosite.Request{
		ID:                rs.Request,
		RequestedAt:       rs.RequestedAt,
		Client:            c,
		RequestedScope:    fosite.Arguments(stringsx.Splitx(rs.Scopes, "|")),
		GrantedScope:      fosite.Arguments(stringsx.Splitx(rs.GrantedScope, "|")),
		RequestedAudience: fosite.Arguments(stringsx.Splitx(rs.RequestedAudience, "|")),
		GrantedAudience:   fosite.Arguments(stringsx.Splitx(rs.GrantedAudience, "|")),
		Form:              val,
		Session:           session,
	}

	return r, nil
}

func (rs *RequesterSql) Value() (driver.Value, error) {
	if rs == nil {
		return nil, nil
	}
	data, _ := json.Marshal(rs)
	return string(data), nil
}

// This method for scanning JSON from json data type in sql
func (rs *RequesterSql) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	var temp RequesterSql
	if err := json.Unmarshal(value.([]byte), &temp); err != nil {
		return err
	}

	*rs = temp
	return nil
}

func toRequesterSql(requester fosite.Requester, signature string) (*RequesterSql, error) {
	subject := requester.GetSession().GetSubject()

	sessionData, err := json.Marshal(requester.GetSession())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &RequesterSql{
		Request:           requester.GetID(),
		Signature:         signature,
		RequestedAt:       requester.GetRequestedAt(),
		Client:            requester.GetClient().GetID(),
		Scopes:            strings.Join([]string(requester.GetRequestedScopes()), ","),
		GrantedScope:      strings.Join([]string(requester.GetGrantedScopes()), ","),
		GrantedAudience:   strings.Join([]string(requester.GetGrantedAudience()), ","),
		RequestedAudience: strings.Join([]string(requester.GetRequestedAudience()), ","),
		Form:              requester.GetRequestForm().Encode(),
		Session:           sessionData,
		Subject:           subject,
	}, nil
}

type AccessTokenSql struct {
	// Signature of token, we don't store whole token for security
	Signature string    `gorm:"signature"`
	Owner     string    `gorm:"owner"`
	RequestId string    `gorm:"request_id"`
	ClientId  string    `gorm:"client_id"`
	Type      string    `gorm:"type"`
	ExpiredAt time.Time `gorm:"expired_at"`
	Requester *RequesterSql
	sdkcm.SQLModel
}

type UserSql struct {
	model.User     `js:",inline"`
	sdkcm.SQLModel `js:",inline"`
}
