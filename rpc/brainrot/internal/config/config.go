package config

import (
	"github.com/meilisearch/meilisearch-go"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	MysqlDataSource string
	Meilisearch     meilisearch.ClientConfig
	MAC             OAUTH2Conf
	APIKey          OAUTH2Conf
	Cache           cache.CacheConf
}

type OAUTH2Conf struct {
	KeyPrefix     string
	RefreshSecret string
	RefreshExpire int64
	Strategy      struct {
		Enable    bool
		Whitelist []string
	}
}
