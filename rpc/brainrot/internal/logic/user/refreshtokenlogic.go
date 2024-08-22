package userlogic

import (
	"context"
	"time"

	"brainrot/gen/pb/brainrot"
	"brainrot/pkg/util"
	"brainrot/rpc/brainrot/internal/svc"
	usermodule "brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Refresh token
func (l *RefreshTokenLogic) RefreshToken(in *brainrot.RefreshTokenRequest) (*brainrot.RefreshTokenResponse, error) {
	if in.RefreshToken == "" {
		return nil, usermodule.ErrInvalidInput
	}

	aes := util.NewAES[svc.RefreshToken](l.svcCtx.Config.MAC.RefreshSecret)
	refreshToken, err := aes.Decrypt(in.RefreshToken)
	if err != nil {
		return nil, usermodule.ErrInvalidRefreshToken
	}

	// TODO: refresh token 应该可以被销毁，这里应该检查是否被销毁

	if refreshToken.ExpireAt < time.Now().Unix() {
		return nil, usermodule.ErrExpiredRefreshToken
	}

	// TODO: 生成了新的凭证后应该销毁旧的凭证

	if l.svcCtx.Config.MAC.Strategy.Enable {
		macresp, err := l.svcCtx.GenMACResponse(refreshToken.UserID)
		if err != nil {
			return nil, err
		}

		macfields := &brainrot.RefreshTokenResponse_MacFields{
			MacFields: &brainrot.MacFields{
				MacId:        macresp.ID,
				MacKey:       macresp.Key,
				MacAlgorithm: macresp.Algorithm,
			},
		}
		return &brainrot.RefreshTokenResponse{
			Auth:               macfields,
			RefreshToken:       macresp.RefreshToken,
			TokenExpire:        l.svcCtx.Config.MAC.KeyExpire,
			RefreshTokenExpire: l.svcCtx.Config.MAC.RefreshExpire,
		}, err
	} else if l.svcCtx.Config.APIKey.Strategy.Enable {
		apiresp, err := l.svcCtx.GenAPIKeyResponse(refreshToken.UserID)
		if err != nil {
			return nil, err
		}

		apikey := &brainrot.RefreshTokenResponse_ApiKey{
			ApiKey: apiresp.Key,
		}
		return &brainrot.RefreshTokenResponse{
			Auth:               apikey,
			RefreshToken:       apiresp.RefreshToken,
			TokenExpire:        l.svcCtx.Config.APIKey.KeyExpire,
			RefreshTokenExpire: l.svcCtx.Config.APIKey.RefreshExpire,
		}, err
	}
	return nil, usermodule.ErrServerError.Wrap("MAC 和 APIKey 策略均未启用")
}
