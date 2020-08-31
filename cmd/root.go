package cmd

import (
	"fmt"
	"github.com/200lab/go-sdk"
	"github.com/200lab/oauth-service/config"
	"github.com/200lab/oauth-service/oauth2"
	"github.com/200lab/oauth-service/oauth2/storage"
	"github.com/200lab/oauth-service/oauth2/usrrepo"
	userStorage "github.com/200lab/oauth-service/oauth2/usrrepo/storage"
	"github.com/200lab/oauth-service/setup"
	"github.com/gin-gonic/gin"
	"github.com/globalsign/mgo"
	"github.com/jinzhu/gorm"
	"github.com/spf13/cobra"
	"log"
	"os"
	"time"
)

type DbConnectionManager interface {
	GetDB() *gorm.DB
	GetRDB() *gorm.DB
}

func oauth2Handlers(db DbConnectionManager, cfg *config.Config) func(engine *gin.Engine) {
	userRepo := usrrepo.New(userStorage.NewSQL(db), cfg)

	return func(engine *gin.Engine) {
		g := engine.Group("oauth2")
		{
			g.GET("/auth", oauth2.AuthHandler)
			g.POST("/auth", oauth2.AuthHandler)
			g.POST("/token", oauth2.AccessTokenHandler)
			g.POST("/introspect", oauth2.IntrospectionHandler)
			users := g.Group("/users")
			{
				users.Use(oauth2.CheckTokenMiddleware)
				users.GET("/:id", oauth2.FindUserHandlerById(userRepo))
				users.POST("", oauth2.CreateUserHandler(userRepo))
				users.PUT("/:id/update", oauth2.UpdateUserHandler(userRepo))
				users.POST("/:id/update", oauth2.UpdateUserHandler(userRepo))
				users.PUT("/:id/change_password", oauth2.ChangePasswordHandler(userRepo))
				users.POST("/:id/change_password", oauth2.ChangePasswordHandler(userRepo))
				users.DELETE("/:id", oauth2.DeleteUserHandler(userRepo))
				users.POST("/:id", oauth2.DeleteUserHandler(userRepo))
			}
		}
	}
}

type mgom struct {
	sc goservice.ServiceContext
}

func (m mgom) GetSession() *mgo.Session {
	return m.sc.MustGet("mdb").(*mgo.Session)
}

var rootCmd = &cobra.Command{
	Use:   "oauth",
	Short: "Start an OAuth Service",
	Run: func(cmd *cobra.Command, args []string) {
		// Init config (need to load config for env for real project or production)
		cfg := config.SystemConfig()

		service := newService()

		//if err := service.Init(); err != nil {
		//	log.Fatal(err)
		//}
		initServiceWithRetry(service, 10)

		db := sqlGorm{service}

		// Init storage and set up if needed
		var store interface{}
		if cfg.StorageType == config.StorageTypeMySQL {
			store = storage.NewSqlStore(db, cfg.GetAES(), cfg.GetSystemSecret())
		} else {
			store = storage.NewMemoryStore()
		}

		// Init OAuth2 Provider with a example config
		oauth2.InitOAuth2Provider(cfg, store)

		// Setup script: Init data for service
		hasher := oauth2.GetHasher()
		st := setup.NewSQL(cfg, db, hasher)
		_ = st.LoadConfig(cfg)

		if st.CanRunInitScript() {
			service.Logger("service").Info("Initializing data...")
			if err := st.Run(); err != nil {
				panic(err)
			}
			service.Logger("service").Info("Initializing data... done")
		}

		// Add OAuth2 handler into gin server
		service.HTTPServer().AddHandler(oauth2Handlers(db, cfg))

		if err := service.Start(); err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	// Add outenv as a sub command
	rootCmd.AddCommand(outEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initServiceWithRetry(s goservice.Service, retry int) {
	var err error

	for i := 1; i <= retry; i++ {
		if err = s.Init(); err != nil {
			time.Sleep(time.Second * 3)
			continue
		} else {
			break
		}
	}

	if err != nil {
		log.Fatal(err)
	}
}

type sqlGorm struct {
	sc goservice.ServiceContext
}

func (m sqlGorm) GetDB() *gorm.DB {
	return m.sc.MustGet("mdb").(*gorm.DB)
}

func (m sqlGorm) GetRDB() *gorm.DB {
	if inter, ok := m.sc.Get("rdb"); ok && inter != nil {
		return inter.(*gorm.DB)
	}

	return nil
}
