package main

import (
	"flag"

	"brainrot/gen/pb/brainrot"
	unaryserverinterceptors "brainrot/pkg/grpcinterceptor"
	"brainrot/pkg/grpcinterceptor/auth/apikey"
	"brainrot/pkg/grpcinterceptor/auth/mac"
	"brainrot/rpc/brainrot/internal/config"
	articleServer "brainrot/rpc/brainrot/internal/server/article"
	pingServer "brainrot/rpc/brainrot/internal/server/ping"
	userServer "brainrot/rpc/brainrot/internal/server/user"
	"brainrot/rpc/brainrot/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/brainrot.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		brainrot.RegisterPingServer(grpcServer, pingServer.NewPingServer(ctx))
		brainrot.RegisterUserServer(grpcServer, userServer.NewUserServer(ctx))
		brainrot.RegisterArticleServer(grpcServer, articleServer.NewArticleServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(unaryserverinterceptors.ErrorToStatusInterceptor)
	if c.MAC.Strategy.Enable {
		s.AddUnaryInterceptors(unaryserverinterceptors.OAuth2AuthorizeInterceptor(
			func() *mac.Authenticator {
				authenticator, err := mac.NewAuthenticator(
					func() *redis.Redis {
						red, err := redis.NewRedis(c.Redis.RedisConf)
						logx.Must(err)
						return red
					}(), c.MAC.KeyPrefix, true, c.MAC.Strategy.Whitelist,
				)
				logx.Must(err)
				return authenticator
			}(),
		))
	} else if c.APIKey.Strategy.Enable {
		s.AddUnaryInterceptors(unaryserverinterceptors.OAuth2AuthorizeInterceptor(
			func() *apikey.Authenticator {
				authenticator, err := apikey.NewAuthenticator(
					func() *redis.Redis {
						red, err := redis.NewRedis(c.Redis.RedisConf)
						logx.Must(err)
						return red
					}(), c.APIKey.KeyPrefix, true, c.APIKey.Strategy.Whitelist,
				)
				logx.Must(err)
				return authenticator
			}(),
		))
	}

	defer s.Stop()

	logx.Debugf("Starting rpc server at %s", c.ListenOn)
	s.Start()
}
