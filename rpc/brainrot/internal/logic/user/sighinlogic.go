package userlogic

import (
	"context"

	"brainrot/gen/pb/brainrot"
	"brainrot/pkg/util/validator"
	"brainrot/rpc/brainrot/internal/svc"
	usermodule "brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type SighInLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSighInLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SighInLogic {
	return &SighInLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Sigh in
func (l *SighInLogic) SighIn(in *brainrot.SighInRequest) (*brainrot.SighInResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, usermodule.ErrLackNecessaryField.Wrap("邮箱或密码为空")
	}

	if !validator.IsEmail(in.Email) {
		return nil, usermodule.ErrInvalidInput.Wrap("邮箱格式不合法")
	}

	usermodel, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, in.Email)
	if err != nil || usermodel == nil {
		return nil, usermodule.ErrDBError.Wrap("根据邮箱查询用户失败，邮箱为：%s", in.Email)
	}

	if usermodel.Password != in.Password {
		return nil, usermodule.ErrInvalidInput.Wrap("密码错误，邮箱为：%s", in.Email)
	}

	resp := &brainrot.SighInResponse{}
	err = copier.Copy(resp, usermodel)
	if err != nil {
		return nil, usermodule.ErrCopierCopy.Wrap("拷贝用户信息失败")
	}

	if l.svcCtx.Config.MAC.Strategy.Enable {
		macresp, err := l.svcCtx.GenMACResponse(int64(usermodel.Id))
		if err != nil {
			return nil, err
		}

		macfields := &brainrot.SighInResponse_MacFields{
			MacFields: &brainrot.MacFields{
				MacId:        macresp.ID,
				MacKey:       macresp.Key,
				MacAlgorithm: macresp.Algorithm,
			},
		}
		resp.Auth = macfields
		resp.RefreshToken = macresp.RefreshToken

	} else if l.svcCtx.Config.APIKey.Strategy.Enable {
		apiresp, err := l.svcCtx.GenAPIKeyResponse(int64(usermodel.Id))
		if err != nil {
			return nil, err
		}

		apikey := &brainrot.SighInResponse_ApiKey{
			ApiKey: apiresp.Key,
		}
		resp.Auth = apikey
		resp.RefreshToken = apiresp.RefreshToken
	} else {
		return nil, usermodule.ErrServerError.Wrap("MAC 和 APIKey 策略均未启用")
	}

	return resp, nil
}
