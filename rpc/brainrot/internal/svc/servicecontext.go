package svc

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/m4n5ter/brainrot/model"
	"github.com/m4n5ter/brainrot/pkg/mac"
	"github.com/m4n5ter/brainrot/pkg/util"
	"github.com/m4n5ter/brainrot/rpc/brainrot/internal/config"
	usermodule "github.com/m4n5ter/brainrot/rpc/brainrot/internal/svc/module/user"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config    config.Config
	UserModel model.UserModel
	Redis     *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	sqlConn := sqlx.NewMysql(c.MysqlDataSource)
	return &ServiceContext{
		Config:    c,
		UserModel: model.NewUserModel(sqlConn),
		Redis: redis.New(c.Redis.Host, func(r *redis.Redis) {
			r.Type = c.Redis.Type
			r.Pass = c.Redis.Pass
		}),
	}
}

func (svcCtx *ServiceContext) GenMACResponse(userid int64) (resp MACResponse, err error) {
	macID, macKey, err := mac.GenerateIDAndKey()
	if err != nil {
		return resp, usermodule.ErrSystemError.Wrap("生成 MAC ID 和 Key 失败")
	}

	htable := fmt.Sprintf("%s%s", svcCtx.Config.MAC.KeyPrefix, macID)
	_, err = svcCtx.Redis.ScriptRun(usermodule.SetScript, []string{htable}, "key", macKey, "userid", strconv.Itoa(int(userid)))
	if err != nil && errors.Is(err, redis.Nil) {
		return resp, usermodule.ErrDBError.Wrap("保存 MAC ID 和 Key 失败，错误为：%v", err)
	}

	refreshTokenMap := map[string]int64{
		"userid":   userid,
		"expireat": time.Now().Unix() + svcCtx.Config.MAC.RefreshExpire,
	}
	aes := util.NewAES[map[string]int64](svcCtx.Config.MAC.Secret)
	refreshToken, err := aes.Encrypt(refreshTokenMap)
	if err != nil {
		return resp, usermodule.ErrSystemError.Wrap("生成 Refresh Token 失败")
	}

	resp.ID = macID
	resp.Key = macKey
	resp.RefreshToken = refreshToken
	resp.Algorithm = "hmac-sha-256"
	return resp, err
}

type MACResponse struct {
	ID           string
	Key          string
	RefreshToken string
	Algorithm    string
}
