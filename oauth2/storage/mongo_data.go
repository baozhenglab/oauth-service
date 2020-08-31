package storage

import (
	"context"
	"encoding/json"
	"github.com/200lab/oauth-service/oauth2/model"
	"github.com/200lab/oauth-service/secure"
	"github.com/globalsign/mgo/bson"
	"github.com/ory/fosite"
	"github.com/ory/go-convenience/stringsx"
	"github.com/pkg/errors"

	"net/url"
	"strings"
	"time"
)

type MgoModel struct {
	PK        bson.ObjectId `bson:"_id,omitempty"`
	IsDeleted bool          `bson:"is_deleted,omitempty"`
	CreatedAt time.Time     `bson:"created_at,omitempty"`
	UpdatedAt time.Time     `bson:"updated_at,omitempty"`
	DeletedAt *time.Time    `bson:"deleted_at,omitempty"`
}

func (md *MgoModel) PrepareForInsert() {
	md.PK = bson.NewObjectId()
	md.CreatedAt = time.Now().UTC()
	md.UpdatedAt = time.Now().UTC()
	md.IsDeleted = false
}

type ClientMongo struct {
	ID                string   `bson:"id"`
	Name              string   `bson:"client_name"`
	Secret            string   `bson:"client_secret"`
	RedirectURIs      []string `bson:"redirect_uris"`
	GrantTypes        []string `bson:"grant_types"`
	ResponseTypes     []string `bson:"response_types"`
	Scope             string   `bson:"scope"`
	Audience          []string `bson:"audiences"`
	OwnerID           string   `bson:"owner_id"`
	PolicyURI         string   `bson:"policy_uri"`
	TermsOfServiceURI string   `bson:"tos_uri"`
	ClientURI         string   `bson:"client_uri"`
	LogoURI           string   `bson:"logo_uri"`
	Contacts          []string `bson:"contacts"`
	SecretExpiresAt   int      `bson:"client_secret_expires_at"`
	MgoModel          `bson:",inline"`
}

func (cm *ClientMongo) toClient() *model.Client {
	c := &model.Client{
		ClientID:          cm.ID,
		Name:              cm.Name,
		Secret:            cm.Secret,
		Audience:          cm.Audience,
		RedirectURIs:      cm.RedirectURIs,
		GrantTypes:        cm.GrantTypes,
		ResponseTypes:     cm.ResponseTypes,
		Scope:             cm.Scope,
		Owner:             cm.OwnerID,
		PolicyURI:         cm.PolicyURI,
		TermsOfServiceURI: cm.TermsOfServiceURI,
		ClientURI:         cm.ClientURI,
		LogoURI:           cm.LogoURI,
		Contacts:          cm.Contacts,
		CreatedAt:         cm.CreatedAt,
		UpdatedAt:         cm.UpdatedAt,
	}

	return c
}

func toClientMongo(c *model.Client) *ClientMongo {
	cm := &ClientMongo{
		ID:                c.ClientID,
		Name:              c.Name,
		Secret:            c.Secret,
		Audience:          c.Audience,
		RedirectURIs:      c.RedirectURIs,
		GrantTypes:        c.GrantTypes,
		ResponseTypes:     c.ResponseTypes,
		Scope:             c.Scope,
		OwnerID:           c.Owner,
		PolicyURI:         c.PolicyURI,
		TermsOfServiceURI: c.TermsOfServiceURI,
		ClientURI:         c.ClientURI,
		LogoURI:           c.LogoURI,
		Contacts:          c.Contacts,
		MgoModel: MgoModel{
			CreatedAt: c.CreatedAt,
			UpdatedAt: c.UpdatedAt,
		},
	}

	return cm
}

type RequesterMongo struct {
	Signature         string      `bson:"signature"`
	Request           string      `bson:"request_id"`
	RequestedAt       time.Time   `bson:"requested_at"`
	Client            string      `bson:"client_id"`
	Scopes            string      `bson:"scope"`
	GrantedScope      string      `bson:"granted_scope"`
	RequestedAudience string      `bson:"requested_audience"`
	GrantedAudience   string      `bson:"granted_audience"`
	Form              string      `bson:"form_data"`
	Subject           string      `bson:"subject"`
	Session           bson.Binary `bson:"session_data"`
	*MgoModel         `bson:",inline"`
}

func (rm *RequesterMongo) toRequester(_ *secure.AES, session fosite.Session, clm ClientManager) (*fosite.Request, error) {
	if session != nil {
		if err := json.Unmarshal(rm.Session.Data, session); err != nil {
			return nil, errors.WithStack(err)
		}
	}

	c, err := clm.GetClient(context.Background(), rm.Client)
	if err != nil {
		return nil, err
	}

	val, err := url.ParseQuery(rm.Form)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	r := &fosite.Request{
		ID:                rm.Request,
		RequestedAt:       rm.RequestedAt,
		Client:            c,
		RequestedScope:    fosite.Arguments(stringsx.Splitx(rm.Scopes, "|")),
		GrantedScope:      fosite.Arguments(stringsx.Splitx(rm.GrantedScope, "|")),
		RequestedAudience: fosite.Arguments(stringsx.Splitx(rm.RequestedAudience, "|")),
		GrantedAudience:   fosite.Arguments(stringsx.Splitx(rm.GrantedAudience, "|")),
		Form:              val,
		Session:           session,
	}

	return r, nil
}

func toRequesterMongo(requester fosite.Requester, signature string) (*RequesterMongo, error) {
	subject := requester.GetSession().GetSubject()

	sessionData, err := json.Marshal(requester.GetSession())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &RequesterMongo{
		Request:           requester.GetID(),
		Signature:         signature,
		RequestedAt:       requester.GetRequestedAt(),
		Client:            requester.GetClient().GetID(),
		Scopes:            strings.Join([]string(requester.GetRequestedScopes()), "|"),
		GrantedScope:      strings.Join([]string(requester.GetGrantedScopes()), "|"),
		GrantedAudience:   strings.Join([]string(requester.GetGrantedAudience()), "|"),
		RequestedAudience: strings.Join([]string(requester.GetRequestedAudience()), "|"),
		Form:              requester.GetRequestForm().Encode(),
		Session:           bson.Binary{Data: sessionData},
		Subject:           subject,
	}, nil
}

type AccessTokenMongo struct {
	// Signature of token, we don't store whole token for security
	Signature string    `bson:"signature"`
	Owner     string    `bson:"owner"`
	RequestID string    `bson:"request_id"`
	ClientID  string    `bson:"client_id"`
	Type      string    `bson:"type"`
	ExpiredAt time.Time `bson:"expired_at"`
	Requester *RequesterMongo
	MgoModel  `bson:",inline"`
}

type UserMongo struct {
	model.User `bson:",inline"`
	MgoModel   `bson:",inline"`
}
