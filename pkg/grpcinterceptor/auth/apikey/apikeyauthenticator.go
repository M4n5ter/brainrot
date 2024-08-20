package apikey

import (
	"context"
	"fmt"
	"strings"
	"time"

	"brainrot/pkg/grpcinterceptor/auth"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	defaultExpiration    = 5 * time.Minute
	missingMetadata      = "missing metadata"
	invalidAuthorization = "invalid authorization"
)

// An Authenticator is used to authenticate the rpc requests.
type Authenticator struct {
	store     *redis.Redis
	keyPrefix string
	cache     *collection.Cache
	strict    bool
	whitelist []string
}

// NewAuthenticator returns an Authenticator.
func NewAuthenticator(store *redis.Redis, keyPrefix string, strict bool, whitelist []string) (*Authenticator, error) {
	cache, err := collection.NewCache(defaultExpiration)
	if err != nil {
		return nil, err
	}

	return &Authenticator{
		store,
		keyPrefix,
		cache,
		strict,
		whitelist,
	}, nil
}

// Authenticate authenticates the given ctx.
func (a *Authenticator) Authenticate(ctx context.Context) (metadata.MD, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, missingMetadata)
	}

	urls := md[auth.HTTPURL]

	if len(urls) == 0 {
		return nil, status.Error(codes.Unauthenticated, missingMetadata)
	}

	url := urls[0]

	if len(url) == 0 {
		return nil, status.Error(codes.Unauthenticated, missingMetadata)
	}

	// 如果在白名单中，直接通过
	for _, v := range a.whitelist {
		if strings.HasPrefix(url, v) {
			return nil, nil
		}
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, missingMetadata)
	}

	token := tokens[0]
	if !strings.HasPrefix(token, "Bearer ") {
		return nil, status.Error(codes.Unauthenticated, "")
	}

	apikey := token[7:]

	userid, err := a.validate(apikey)
	if err != nil {
		return nil, err
	}

	return metadata.Pairs("userid", userid), err
}

func (a *Authenticator) validate(apikey string) (userid string, err error) {
	htable := fmt.Sprintf("%s%s", a.keyPrefix, apikey)

	expectedUserID, err := a.cache.Take(apikey, func() (any, error) {
		return a.store.Hget(htable, "userid")
	})
	if err != nil {
		if a.strict {
			return "", status.Error(codes.Internal, err.Error())
		}

		return "", nil
	}
	if expectedUserID == nil {
		return "", status.Error(codes.Unauthenticated, "userid not found")
	}

	userid = expectedUserID.(string)
	return userid, nil
}
