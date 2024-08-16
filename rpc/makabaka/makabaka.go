package main

import (
	"flag"

	"github.com/m4n5ter/makabaka/pb/makabaka"
	unaryserverinterceptors "github.com/m4n5ter/makabaka/pkg/grpcinterceptor"
	"github.com/m4n5ter/makabaka/pkg/grpcinterceptor/auth"
	"github.com/m4n5ter/makabaka/rpc/makabaka/internal/config"
	pingServer "github.com/m4n5ter/makabaka/rpc/makabaka/internal/server/ping"
	userServer "github.com/m4n5ter/makabaka/rpc/makabaka/internal/server/user"
	"github.com/m4n5ter/makabaka/rpc/makabaka/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/makabaka.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		makabaka.RegisterPingServer(grpcServer, pingServer.NewPingServer(ctx))
		makabaka.RegisterUserServer(grpcServer, userServer.NewUserServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	s.AddUnaryInterceptors(unaryserverinterceptors.ErrorToStatusInterceptor)
	if c.MAC.Strategy.Enable {
		s.AddUnaryInterceptors(unaryserverinterceptors.OAuth2MACAuthorizeInterceptor(
			func() *auth.Authenticator {
				authenticator, err := auth.NewAuthenticator(
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
	}

	defer s.Stop()

	logx.Debugf("Starting rpc server at %s", c.ListenOn)
	s.Start()
}
