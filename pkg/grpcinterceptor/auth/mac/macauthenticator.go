package mac

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"brainrot/pkg/grpcinterceptor/auth"
	"brainrot/pkg/mac"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	cache, err := collection.NewCache(auth.DefaultExpiration)
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
		return nil, status.Error(codes.Unauthenticated, auth.MissingMetadata)
	}

	methods, urls, hostnames, ports := md[auth.HTTPMethod], md[auth.HTTPURL], md[auth.HTTPHostname], md[auth.HTTPPort]

	if len(methods) == 0 || len(urls) == 0 || len(hostnames) == 0 || len(ports) == 0 {
		return nil, status.Error(codes.Unauthenticated, auth.MissingMetadata)
	}

	method, url, hostname, port := methods[0], urls[0], hostnames[0], ports[0]

	if len(method) == 0 || len(url) == 0 || len(hostname) == 0 || len(port) == 0 {
		return nil, status.Error(codes.Unauthenticated, auth.MissingMetadata)
	}

	// 如果在白名单中，直接通过
	for _, v := range a.whitelist {
		if strings.HasPrefix(url, v) {
			return nil, nil
		}
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 {
		return nil, status.Error(codes.Unauthenticated, auth.MissingMetadata)
	}

	token := tokens[0]
	if !strings.HasPrefix(token, "MAC ") {
		return nil, status.Error(codes.Unauthenticated, "")
	}

	var id, ts, nonce, ext, macstr string
	parts := strings.Split(token[len("MAC "):], ",")
	// The count of header attributes must be 4 or 5.
	if len(parts) != 4 && len(parts) != 5 {
		return nil, status.Error(codes.Unauthenticated, auth.InvalidAuthorization)
	}

	for _, part := range parts {
		trimed := strings.TrimSpace(part)
		i := strings.Index(trimed, "=")
		if i == -1 {
			return nil, status.Error(codes.Unauthenticated, auth.InvalidAuthorization)
		}
		k, v := trimed[:i], strings.TrimFunc(trimed[i+1:], func(r rune) bool {
			return r == '"'
		})

		switch k {
		case "id":
			id = v
		case "ts":
			ts = v
		case "nonce":
			nonce = v
		case "ext":
			ext = v
		case "mac":
			macstr = v
		default:
			return nil, status.Error(codes.Unauthenticated, auth.InvalidAuthorization)
		}
	}

	if len(id) == 0 || len(ts) == 0 || len(nonce) == 0 || len(macstr) == 0 {
		return nil, status.Error(codes.Unauthenticated, auth.InvalidAuthorization)
	}

	userid, err := a.validate(mac.NewRequest(
		id, ts, nonce, ext, macstr,
	), method, url, hostname, port)
	if err != nil {
		return nil, err
	}

	return metadata.Pairs("userid", userid), err
}

func (a *Authenticator) validate(macreq mac.Request, method, url, hostname, port string) (userid string, err error) {
	// 验证时间戳
	tsSec, err := strconv.Atoi(macreq.TS)
	if err != nil {
		return "", status.Error(codes.Unauthenticated, "invalid timestamp")
	}

	ts := time.Unix(int64(tsSec), 0).Unix()
	now := time.Now().Unix()
	if now-ts > 60 {
		return "", status.Error(codes.Unauthenticated, "timestamp expired")
	}
	if ts-now > 3 {
		return "", status.Error(codes.Unauthenticated, "timestamp invalid")
	}

	// 验证 id 是否存在，存在的话取出对应的 key
	htable := fmt.Sprintf("%s%s", a.keyPrefix, macreq.ID)

	expectKey, err := a.cache.Take(macreq.ID, func() (any, error) {
		return a.store.Hget(htable, "key")
	})
	if err != nil {
		if a.strict {
			return "", status.Error(codes.Internal, err.Error())
		}

		return "", nil
	}
	if expectKey == nil {
		return "", status.Error(codes.Unauthenticated, "mac key not found")
	}

	// 计算 MAC
	reqstr := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n\n", macreq.TS, macreq.Nonce, method, url, hostname, port)
	h := hmac.New(sha256.New, []byte(expectKey.(string)))
	h.Write([]byte(reqstr))
	if calculatedMAC := base64.StdEncoding.EncodeToString(h.Sum(nil)); calculatedMAC != macreq.MAC {
		return "", status.Error(codes.Unauthenticated, "invalid mac")
	}

	// 验证 MAC ID, timestamp, nonce 组合的唯一性
	combination := fmt.Sprintf("%s:%s", macreq.TS, macreq.Nonce)

	existCombination, err := a.cache.Take(fmt.Sprintf("%s:%s", macreq.ID, combination), func() (any, error) {
		return a.store.Hsetnx(htable, combination, "1")
	})
	if err != nil {
		if a.strict {
			return "", status.Error(codes.Internal, err.Error())
		}

		return "", nil
	}
	if exist := existCombination.(bool); !exist {
		return "", status.Error(codes.Unauthenticated, "repeated nonce")
	}

	// 验证通过取出 userid
	userid, err = a.store.Hget(htable, "userid")
	if err != nil {
		if a.strict {
			return "", status.Error(codes.Internal, err.Error())
		}

		return "", nil
	}
	// 到这一步 userid 一定存在
	return userid, nil
}
