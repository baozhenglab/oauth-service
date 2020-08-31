package cmd

import (
	"github.com/200lab/go-sdk"
	"github.com/200lab/go-sdk/plugin/storage/sdkgorm"
	"github.com/200lab/go-sdk/plugin/storage/sdkredis"
)

var (
	serviceName = "oauth-service"
	version     = "1.0.0"
)

func newService() goservice.Service {
	s := goservice.New(
		goservice.WithName(serviceName),
		goservice.WithVersion(version),
		goservice.WithInitRunnable(sdkgorm.NewGormDB("mdb", "mdb")),
		goservice.WithInitRunnable(sdkredis.NewRedisDB("redis", "redis")),
	)

	return s
}
