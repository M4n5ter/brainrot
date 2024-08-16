package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	MysqlDataSource string
	MAC             struct {
		KeyPrefix     string
		Secret        string
		RefreshExpire int64
		Strategy      struct {
			Enable    bool
			Whitelist []string
		}
	}
}
