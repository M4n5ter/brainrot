package userlogic

import (
	"context"
	"encoding/hex"
	"errors"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/pkg/util"
	"brainrot/pkg/util/validator"
	"brainrot/rpc/brainrot/internal/svc"
	usermodule "brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type SignInLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSignInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SignInLogic {
	return &SignInLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Sign in
func (l *SignInLogic) SignIn(in *brainrot.SignInRequest) (*brainrot.SignInResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, usermodule.ErrLackNecessaryField.Wrap("邮箱或密码为空")
	}

	if !validator.IsEmail(in.Email) {
		return nil, usermodule.ErrInvalidInput.Wrap("邮箱格式不合法")
	}

	modeluser, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, in.Email)
	if err != nil || modeluser == nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, usermodule.ErrDBUserNotFound
		}
		return nil, usermodule.ErrDBError.Wrap("根据邮箱查询用户失败，邮箱为：%s，错误为：%v", in.Email, err)
	}

	if modeluser.Password != hex.EncodeToString(util.HashWithSalt(util.TobrainrotBytes(in.Password), nil)) {
		return nil, usermodule.ErrInvalidInput.Wrap("密码错误，邮箱为：%s", in.Email)
	}

	resp := &brainrot.SignInResponse{}
	err = copier.Copy(resp, modeluser)
	if err != nil {
		return nil, usermodule.ErrCopierCopy.Wrap("拷贝用户信息失败")
	}

	if l.svcCtx.Config.MAC.Strategy.Enable {
		macresp, err := l.svcCtx.GenMACResponse(int64(modeluser.Id))
		if err != nil {
			return nil, err
		}

		macfields := &brainrot.SignInResponse_MacFields{
			MacFields: &brainrot.MacFields{
				MacId:        macresp.ID,
				MacKey:       macresp.Key,
				MacAlgorithm: macresp.Algorithm,
			},
		}
		resp.Auth = macfields
		resp.RefreshToken = macresp.RefreshToken
		resp.TokenExpire = l.svcCtx.Config.MAC.KeyExpire
		resp.RefreshTokenExpire = l.svcCtx.Config.MAC.RefreshExpire

	} else if l.svcCtx.Config.APIKey.Strategy.Enable {
		apiresp, err := l.svcCtx.GenAPIKeyResponse(int64(modeluser.Id))
		if err != nil {
			return nil, err
		}

		apikey := &brainrot.SignInResponse_ApiKey{
			ApiKey: apiresp.Key,
		}
		resp.Auth = apikey
		resp.RefreshToken = apiresp.RefreshToken
		resp.TokenExpire = l.svcCtx.Config.APIKey.KeyExpire
		resp.RefreshTokenExpire = l.svcCtx.Config.APIKey.RefreshExpire
	} else {
		return nil, usermodule.ErrServerError.Wrap("MAC 和 APIKey 策略均未启用")
	}

	return resp, nil
}
