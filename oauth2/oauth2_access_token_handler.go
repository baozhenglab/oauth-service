package oauth2

import (
	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"log"
	"time"
)

func AccessTokenHandler(c *gin.Context) {
	// This context will be passed to all methods.
	ctx := fosite.NewContext()

	// Create an empty session object which will be passed to the request handlers
	mySessionData := newSession(c.PostForm("username"))
	// This will create an access request object and iterate through the registered TokenEndpointHandlers to validate the request.
	accessRequest, err := oauth2.NewAccessRequest(ctx, c.Request, mySessionData)

	// Catch any errors, e.g.:
	// * unknown client
	// * invalid redirect
	// * ...
	if err != nil {
		log.Printf("Error occurred in NewAccessRequest: %+v", err)
		oauth2.WriteAccessError(c.Writer, accessRequest, err)
		return
	}

	// If this is a client_credentials grant, grant all scopes the client is allowed to perform.
	if accessRequest.GetGrantTypes().Exact("client_credentials") {
		accessRequest.GetSession().SetExpiresAt(fosite.AccessToken, time.Now().Add(time.Hour*24*365))
		for _, scope := range accessRequest.GetRequestedScopes() {
			if fosite.HierarchicScopeStrategy(accessRequest.GetClient().GetScopes(), scope) {
				accessRequest.GrantScope(scope)
			}
		}
	}

	if accessRequest.GetGrantTypes().Exact("password") {
		accessRequest.GrantScope("offline")
	}

	if accessRequest.GetGrantTypes().Exact("refresh_token") {
		accessRequest.GetSession().SetExpiresAt(fosite.AccessToken, time.Now().Add(time.Hour*24*45))
	}

	// Next we create a response for the access request. Again, we iterate through the TokenEndpointHandlers
	// and aggregate the result in response.
	response, err := oauth2.NewAccessResponse(ctx, accessRequest)
	if err != nil {
		log.Printf("Error occurred in NewAccessResponse: %+v", err)
		oauth2.WriteAccessError(c.Writer, accessRequest, err)
		return
	}

	// All done, send the response.
	oauth2.WriteAccessResponse(c.Writer, accessRequest, response)
}
