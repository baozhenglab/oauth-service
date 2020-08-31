package ginhandler

import (
	"github.com/baozhenglab/oauth-service/config"
	"github.com/baozhenglab/oauth-service/oauth2"
	"github.com/baozhenglab/oauth-service/oauth2/usrrepo"
	userStorage "github.com/baozhenglab/oauth-service/oauth2/usrrepo/storage"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type DbConnectionManager interface {
	GetDB() *gorm.DB
	GetRDB() *gorm.DB
}

func Oauth2Handlers(db DbConnectionManager, cfg *config.Config) func(engine *gin.Engine) {
	userRepo := usrrepo.New(userStorage.NewSQL(db), cfg)

	return func(engine *gin.Engine) {
		g := engine.Group("oauth2")
		{
			g.GET("/auth", oauth2.AuthHandler)
			g.POST("/auth", oauth2.AuthHandler)
			g.POST("/token", oauth2.AccessTokenHandler)
			g.POST("/introspect", oauth2.IntrospectionHandler)
			g.POST("/find-user", oauth2.FindUserHandler(userRepo))

			g.POST("/generate-otp", oauth2.CheckTokenMiddleware, oauth2.GenerateOTP(userRepo))
			g.POST("/login-otp", oauth2.CheckTokenMiddleware, oauth2.LoginWithOTP(userRepo))
			g.POST("/login", oauth2.CheckTokenMiddleware, oauth2.LoginOtherCredential(userRepo))

			users := g.Group("/users")
			{
				users.Use(oauth2.CheckTokenMiddleware)
				users.GET("/:id", oauth2.FindUserHandlerById(userRepo))
				users.POST("", oauth2.CreateUserHandler(userRepo))
				users.PUT("/:id/update", oauth2.UpdateUserHandler(userRepo))
				users.POST("/:id/update", oauth2.UpdateUserHandler(userRepo))
				users.PUT("/:id/change-password", oauth2.ChangePasswordHandler(userRepo))
				users.POST("/:id/change-password", oauth2.ChangePasswordHandler(userRepo))
				users.POST("/:id/set-username-password", oauth2.SetUsernamePasswordHandler(userRepo))
				users.DELETE("/:id", oauth2.DeleteUserHandler(userRepo))
				users.POST("/:id", oauth2.DeleteUserHandler(userRepo))
			}
		}
	}
}
