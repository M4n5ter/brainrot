package svc

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"brainrot/model"
	"brainrot/pkg/mac"
	"brainrot/pkg/util"
	"brainrot/rpc/brainrot/internal/config"
	usermodule "brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/meilisearch/meilisearch-go"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config       config.Config
	UserModel    model.UserModel
	ArticleModel model.ArticleModel
	TagModel     model.TagModel
	CommentModel model.CommentModel
	Redis        *redis.Redis
	Meili        *meilisearch.Client
	S3           *minio.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.MysqlDataSource)

	s3, err := minio.New(c.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(c.S3.AccessKeyID, c.S3.SecretAccessKey, ""),
		Secure: c.S3.UseSSL,
		Region: c.S3.Region,
	})
	logx.Must(err)

	return &ServiceContext{
		Config:       c,
		UserModel:    model.NewUserModel(sqlConn, c.Cache),
		ArticleModel: model.NewArticleModel(sqlConn, c.Cache),
		TagModel:     model.NewTagModel(sqlConn, c.Cache),
		CommentModel: model.NewCommentModel(sqlConn, c.Cache),
		Redis: redis.New(c.Redis.Host, func(r *redis.Redis) {
			r.Type = c.Redis.Type
			r.Pass = c.Redis.Pass
		}),
		Meili: meilisearch.NewClient(c.Meilisearch.ToClientConfig()),
		S3:    s3,
	}
}

func (svcCtx *ServiceContext) GenMACResponse(userid int64) (resp MACResponse, err error) {
	macID, macKey, err := mac.GenerateIDAndKey()
	if err != nil {
		return resp, usermodule.ErrSystemError.Wrap("生成 MAC ID 和 Key 失败")
	}

	htable := fmt.Sprintf("%s%s", svcCtx.Config.MAC.KeyPrefix, macID)
	_, err = svcCtx.Redis.ScriptRun(usermodule.SetScript, []string{htable, strconv.Itoa(int(svcCtx.Config.MAC.KeyExpire))}, "key", macKey, "userid", strconv.Itoa(int(userid)))
	if err != nil && !errors.Is(err, redis.Nil) {
		return resp, usermodule.ErrDBError.Wrap("保存 MAC ID 和 Key 失败，错误为：%v", err)
	}

	refreshExpire := defaultRefreshExpire
	if svcCtx.Config.APIKey.RefreshExpire != 0 {
		refreshExpire = svcCtx.Config.APIKey.RefreshExpire
	}
	refreshToken := RefreshToken{
		UserID:   userid,
		ExpireAt: time.Now().Unix() + refreshExpire,
	}
	aes := util.NewAES[RefreshToken](svcCtx.Config.MAC.RefreshSecret)
	token, err := aes.Encrypt(refreshToken)
	if err != nil {
		return resp, usermodule.ErrSystemError.Wrap("生成 Refresh Token 失败")
	}

	resp.ID = macID
	resp.Key = macKey
	resp.RefreshToken = token
	resp.Algorithm = "hmac-sha-256"
	return resp, err
}

type MACResponse struct {
	ID           string
	Key          string
	RefreshToken string
	Algorithm    string
}

func (svcCtx *ServiceContext) GenAPIKeyResponse(userid int64) (resp APIKeyResponse, err error) {
	key, err := util.GenerateRandomHexString()
	if err != nil {
		return APIKeyResponse{}, usermodule.ErrSystemError.Wrap("生成 API Key 失败：%v", err)
	}

	htable := fmt.Sprintf("%s%s", svcCtx.Config.APIKey.KeyPrefix, key)
	_, err = svcCtx.Redis.ScriptRun(usermodule.SetScript, []string{htable, strconv.Itoa(int(svcCtx.Config.APIKey.KeyExpire))}, "userid", strconv.Itoa(int(userid)))
	if err != nil && !errors.Is(err, redis.Nil) {
		return resp, usermodule.ErrDBError.Wrap("保存 API Key 失败，错误为：%v", err)
	}

	refreshExpire := defaultRefreshExpire
	if svcCtx.Config.APIKey.RefreshExpire != 0 {
		refreshExpire = svcCtx.Config.APIKey.RefreshExpire
	}
	refreshToken := RefreshToken{
		UserID:   userid,
		ExpireAt: time.Now().Unix() + refreshExpire,
	}
	aes := util.NewAES[RefreshToken](svcCtx.Config.APIKey.RefreshSecret)
	token, err := aes.Encrypt(refreshToken)
	if err != nil {
		return resp, usermodule.ErrSystemError.Wrap("生成 Refresh Token 失败：%v", err)
	}

	resp.Key = key
	resp.RefreshToken = token
	return resp, nil
}

type APIKeyResponse struct {
	Key          string
	RefreshToken string
}

type RefreshToken struct {
	UserID   int64 `json:"userid"`
	ExpireAt int64 `json:"expire_at"`
}

const defaultRefreshExpire int64 = 907200
