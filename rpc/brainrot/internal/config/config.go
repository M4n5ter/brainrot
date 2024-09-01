package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	MysqlDataSource string
	Meilisearch     MeiliConf
	S3              S3Conf
	MAC             OAUTH2Conf `json:",optional"` //nolint:staticcheck
	APIKey          OAUTH2Conf `json:",optional"` //nolint:staticcheck
	Cache           cache.CacheConf
}

type OAUTH2Conf struct {
	KeyPrefix string
	// Default 604800 seconds(7 days)
	KeyExpire     int64 `json:",default=604800"` //nolint:staticcheck
	RefreshSecret string
	RefreshExpire int64 `json:",default=604800"` //nolint:staticcheck
	Strategy      struct {
		Enable    bool
		Whitelist []string
	}
}

type MeiliConf struct {
	// Host is the host of your Meilisearch database
	// Example: 'http://localhost:7700'
	Host string

	// APIKey is optional
	APIKey string `json:",optional"` //nolint:staticcheck
}

type S3Conf struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	PublicBucket    string
	PrivateBucket   string
	UseSSL          bool `json:",default=false"` //nolint:staticcheck
}
