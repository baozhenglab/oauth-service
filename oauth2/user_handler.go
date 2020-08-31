package oauth2

import (
	"context"
	"log"
	"net/http"
	"strings"

	//"github.com/200lab/oauth-service/common"
	sdkcmn "github.com/baozhenglab/go-sdk/sdkcm"
	"github.com/baozhenglab/oauth-service/oauth2/model"
	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
)

type UserRepo interface {
	Find(ctx context.Context, filter *model.UserFilter) (u *model.User, err error)
	Create(ctx context.Context, user *model.User) (u *model.User, err error)
	CreateWithFacebook(ctx context.Context, fbId, email, clientId string) (u *model.User, err error)
	CreateWithAccountKit(ctx context.Context, akId, email, prefix, phone, clientId string) (u *model.User, err error)
	CreateWithGmail(ctx context.Context, email, clientId string) (u *model.User, err error)
	CreateWithApple(ctx context.Context, appleId, email, clientId string) (u *model.User, err error)
	ChangePassword(ctx context.Context, clientId, uid, oldPass, newPass string) error
	UpdateUser(ctx context.Context, usrUpdate *model.UserUpdate) (*model.User, error)
	SetUsernamePassword(ctx context.Context, user *model.CredentialAndPassword) error
	GenerateOTP(ctx context.Context, userFilter *model.UserFilter) (string, error)
	LoginWithOTP(ctx context.Context, userFilter *model.UserFilter) (*model.User, error)
	LoginWithOtherCredentialAndPassword(ctx context.Context, credential *model.CredentialAndPassword) (*model.User, error)
	Delete(ctx context.Context, clientId, uid string) error
}

func CheckTokenMiddleware(c *gin.Context) {
	r := c.Request
	token := fosite.AccessTokenFromRequest(r)

	session := newSession("introspect")
	_, ar, err := oauth2.(*fosite.Fosite).IntrospectToken(context.Background(), token, fosite.AccessToken, session.Clone())

	if err != nil {
		_ = c.AbortWithError(http.StatusUnauthorized, err)
		return
	}

	c.Set("client_id", ar.GetClient().GetID())
	c.Set("client", ar.GetClient())
	c.Next()
}

func FindUserHandlerById(ur UserRepo) func(*gin.Context) {
	return func(c *gin.Context) {
		uid := c.Param("id")
		res, err := ur.Find(c.Request.Context(), &model.UserFilter{UserId: &uid})

		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		c.JSON(http.StatusOK, sdkcmn.SimpleSuccessResponse(res))
	}
}

func FindUserHandler(ur UserRepo) func(*gin.Context) {
	return func(c *gin.Context) {
		var p model.UserFilter
		if err := c.ShouldBind(&p); err != nil {
			cErr := sdkcmn.ErrInvalidRequest(err)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		res, err := ur.Find(c.Request.Context(), &p)

		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		c.JSON(http.StatusOK, sdkcmn.SimpleSuccessResponse(res))
	}
}

// These handles no belong to original oauth2 protocol
// Create a new user
func CreateUserHandler(ur UserRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		switch c.DefaultQuery("type", "direct") {
		case "direct":
			createUserDirectly(ur, c)
			return
		case "facebook":
			createUserByFacebook(ur, c)
			return
		case "account-kit":
			createUserByAccountKit(ur, c)
			return
		case "gmail":
			createUserByGmail(ur, c)
			return
		case "apple":
			createUserByApple(ur, c)
			return
		}
	}
}

func createUserByFacebook(ur UserRepo, c *gin.Context) {
	fbId := strings.TrimSpace(c.PostForm("fb_id"))
	email := strings.TrimSpace(c.PostForm("email"))

	cltId, _ := c.Get("client_id")
	clientId := cltId.(string)

	newUser, err := ur.CreateWithFacebook(context.Background(), fbId, email, clientId)
	if err != nil {
		cErr := err.(sdkcmn.AppError)
		c.JSON(cErr.StatusCode, cErr)
		return
	}

	responseUserToken(newUser, c)
}

// Create user directly
func createUserDirectly(ur UserRepo, c *gin.Context) {
	var user model.User

	uname := strings.TrimSpace(c.PostForm("username"))
	if uname != "" {
		user.Username = &uname
	}

	user.Password = strings.TrimSpace(c.PostForm("password"))
	email := c.PostForm("email")
	user.Email = &email

	phone := c.PostForm("phone")
	phonePrefix := c.PostForm("phone_prefix")

	user.Phone = &phone
	user.PhonePrefix = &phonePrefix

	clientId, _ := c.Get("client_id")
	user.ClientId = clientId.(string)

	user.AccountType = model.AccTypeInternal
	newUser, err := ur.Create(context.Background(), &user)
	if err != nil {
		cErr := err.(sdkcmn.AppError)
		c.JSON(cErr.StatusCode, cErr)
		return
	}

	responseUserToken(newUser, c)
}

func createUserByAccountKit(ur UserRepo, c *gin.Context) {
	akId := strings.TrimSpace(c.PostForm("ak_id"))
	email := strings.TrimSpace(c.PostForm("email"))
	prefix := strings.TrimSpace(c.PostForm("phone_prefix"))
	phone := strings.TrimSpace(c.PostForm("phone"))

	cltId, _ := c.Get("client_id")
	clientId := cltId.(string)

	newUser, err := ur.CreateWithAccountKit(context.Background(), akId, email, prefix, phone, clientId)
	if err != nil {
		cErr := err.(sdkcmn.AppError)
		c.JSON(cErr.StatusCode, cErr)
		return
	}

	responseUserToken(newUser, c)
}

func createUserByGmail(ur UserRepo, c *gin.Context) {

	email := strings.TrimSpace(c.PostForm("email"))

	cltId, _ := c.Get("client_id")
	clientId := cltId.(string)

	newUser, err := ur.CreateWithGmail(context.Background(), email, clientId)
	if err != nil {
		cErr := err.(sdkcmn.AppError)
		c.JSON(cErr.StatusCode, cErr)
		return
	}

	responseUserToken(newUser, c)
}

func createUserByApple(ur UserRepo, c *gin.Context) {
	appleId := strings.TrimSpace(c.PostForm("apple_id"))
	email := strings.TrimSpace(c.PostForm("email"))

	cltId, _ := c.Get("client_id")
	clientId := cltId.(string)

	newUser, err := ur.CreateWithApple(context.Background(), appleId, email, clientId)
	if err != nil {
		cErr := err.(sdkcmn.AppError)
		c.JSON(cErr.StatusCode, cErr)
		return
	}

	responseUserToken(newUser, c)
}

// Change password of an user
func ChangePasswordHandler(ur UserRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		cid, _ := c.Get("client_id")
		clientId := cid.(string)

		uid := c.Param("id")

		type param struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		var p param
		if err := c.ShouldBind(&p); err != nil {
			cErr := sdkcmn.ErrInvalidRequest(err)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		err := ur.ChangePassword(context.Background(), clientId, uid, p.OldPassword, p.NewPassword)
		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		c.JSON(http.StatusOK, sdkcmn.SimpleSuccessResponse("ok"))
	}
}

func UpdateUserHandler(ur UserRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		cid, _ := c.Get("client_id")
		clientId := cid.(string)

		uid := c.Param("id")

		var p model.UserUpdate
		if err := c.ShouldBind(&p); err != nil {
			cErr := sdkcmn.ErrInvalidRequest(err)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		p.Id = uid
		p.ClientId = &clientId

		user, err := ur.UpdateUser(c.Request.Context(), &p)
		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		c.JSON(http.StatusOK, sdkcmn.SimpleSuccessResponse(user))
	}
}

// Set usetname & password for an user
func SetUsernamePasswordHandler(ur UserRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		cid, _ := c.Get("client_id")
		clientId := cid.(string)

		uid := c.Param("id")

		var p model.CredentialAndPassword
		if err := c.ShouldBind(&p); err != nil {
			cErr := sdkcmn.ErrInvalidRequest(err)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		p.Id = uid
		p.ClientId = clientId

		err := ur.SetUsernamePassword(c.Request.Context(), &p)
		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		c.JSON(http.StatusOK, sdkcmn.SimpleSuccessResponse("ok"))
	}
}

// Delete user
func DeleteUserHandler(ur UserRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		cid, _ := c.Get("client_id")
		clientId := cid.(string)

		uid := c.Param("id")

		err := ur.Delete(context.Background(), clientId, uid)
		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		c.JSON(http.StatusOK, sdkcmn.SimpleSuccessResponse("ok"))
	}
}

func responseUserToken(user *model.User, c *gin.Context) {
	client := c.MustGet("client").(fosite.Client)

	session := newSession(user.UserId)

	email := ""
	if user.Email != nil {
		email = *user.Email
	}
	session.SetUserEmail(email)
	session.SetUserID(user.UserId)

	ar := fosite.NewAccessRequest(session)
	ar.SetRequestedScopes([]string{"root", "offline"})
	ar.GrantedScope = []string{"offline"}
	ar.GrantTypes = []string{"password"}
	ar.Client = client

	response, err := oauth2.NewAccessResponse(context.Background(), ar)

	if err != nil {
		log.Printf("Error occurred in NewAccessResponse: %+v", err)
		oauth2.WriteAccessError(c.Writer, ar, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"oauth_id":              user.UserId,
		"access_token":          response.GetAccessToken(),
		"refresh_token":         response.GetExtra("refresh_token"),
		"expires_in":            response.GetExtra("expires_in"),
		"is_new":                user.IsNew,
		"has_username_password": user.HasUsernamePassword,
	})
}

func LoginOtherCredential(ur UserRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		cid, _ := c.Get("client_id")
		clientId := cid.(string)

		var p model.CredentialAndPassword
		if err := c.ShouldBind(&p); err != nil {
			cErr := sdkcmn.ErrInvalidRequest(err)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		p.ClientId = clientId

		user, err := ur.LoginWithOtherCredentialAndPassword(c.Request.Context(), &p)

		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		responseUserToken(user, c)
	}
}

func LoginWithOTP(ur UserRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		cid, _ := c.Get("client_id")
		clientId := cid.(string)

		var p model.UserFilter
		if err := c.ShouldBind(&p); err != nil {
			cErr := sdkcmn.ErrInvalidRequest(err)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		p.ClientId = &clientId

		user, err := ur.LoginWithOTP(c.Request.Context(), &p)

		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		responseUserToken(user, c)
	}
}

func GenerateOTP(ur UserRepo) func(c *gin.Context) {
	return func(c *gin.Context) {
		cid, _ := c.Get("client_id")
		clientId := cid.(string)

		var p model.UserFilter
		if err := c.ShouldBind(&p); err != nil {
			cErr := sdkcmn.ErrInvalidRequest(err)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		p.ClientId = &clientId

		otp, err := ur.GenerateOTP(c.Request.Context(), &p)

		if err != nil {
			cErr := err.(sdkcmn.AppError)
			c.JSON(cErr.StatusCode, cErr)
			return
		}

		c.JSON(http.StatusOK, sdkcmn.SimpleSuccessResponse(otp))
	}
}
