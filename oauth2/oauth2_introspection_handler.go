package oauth2

import (
	"encoding/json"
	"github.com/200lab/go-sdk/sdkcm"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
)

func IntrospectionHandler(c *gin.Context) {
	ctx := fosite.NewContext()
	mySessionData := newSession("introspect")
	ir, err := oauth2.NewIntrospectionRequest(ctx, c.Request, mySessionData)
	if err != nil {
		//log.Printf("Error occurred in NewAuthorizeRequest: %+v", err)
		//log.Println(err.(*errors.withStack).Cause())
		//oauth2.WriteIntrospectionError(c.Writer, err)
		WriteIntrospectionError(c, err)
		return
	}

	WriteIntrospectionResponse(c.Writer, ir)
}

func WriteIntrospectionResponse(rw http.ResponseWriter, r fosite.IntrospectionResponder) {
	if !r.IsActive() {
		_ = json.NewEncoder(rw).Encode(&struct {
			Active bool `json:"active"`
		}{Active: false})
		return
	}

	expiresAt := int64(0)
	if !r.GetAccessRequester().GetSession().GetExpiresAt(fosite.AccessToken).IsZero() {
		expiresAt = r.GetAccessRequester().GetSession().GetExpiresAt(fosite.AccessToken).Unix()
	}

	type s interface {
		GetUserID() string
		GetEmail() string
	}

	rw.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(rw).Encode(struct {
		Active    bool     `json:"active"`
		ClientID  string   `json:"client_id,omitempty"`
		Scope     string   `json:"scope,omitempty"`
		Audience  []string `json:"aud,omitempty"`
		ExpiresAt int64    `json:"exp,omitempty"`
		IssuedAt  int64    `json:"iat,omitempty"`
		Subject   string   `json:"sub,omitempty"`
		Username  string   `json:"username,omitempty"`
		Email     string   `json:"email"`
		UserId    string   `json:"user_id"`
		// Session is not included per default because it might expose sensitive information.
	}{
		Active:    true,
		ClientID:  r.GetAccessRequester().GetClient().GetID(),
		Scope:     strings.Join(r.GetAccessRequester().GetGrantedScopes(), " "),
		ExpiresAt: expiresAt,
		IssuedAt:  r.GetAccessRequester().GetRequestedAt().Unix(),
		Subject:   r.GetAccessRequester().GetSession().GetSubject(),
		Audience:  r.GetAccessRequester().GetGrantedAudience(),
		Username:  r.GetAccessRequester().GetSession().GetUsername(),
		// Session is not included because it might expose sensitive information.
		Email:  r.GetAccessRequester().GetSession().(s).GetEmail(),
		UserId: r.GetAccessRequester().GetSession().(s).GetUserID(),
	})
}

func WriteIntrospectionError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	//switch errors.Cause(err).Error() {
	//case fosite.ErrInvalidRequest.Error(), fosite.ErrRequestUnauthorized.Error(), fosite.ErrNotFound.Error():
	//
	//	rfcerr := fosite.ErrorToRFC6749Error(err)
	//	c.AbortWithStatusJSON(rfcerr.StatusCode(), rfcerr)
	//	return
	//}
	//
	//c.AbortWithStatusJSON(http.StatusUnauthorized, struct {
	//	Active bool `json:"active"`
	//}{Active: false})

	rfcerr := fosite.ErrorToRFC6749Error(err)

	appErr := sdkcm.NewAppErr(err, rfcerr.StatusCode(), rfcerr.Description).WithCode(rfcerr.Name)
	c.AbortWithStatusJSON(rfcerr.StatusCode(), appErr)
}
