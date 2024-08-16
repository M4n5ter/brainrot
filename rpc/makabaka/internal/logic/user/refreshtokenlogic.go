package userlogic

import (
	"context"
	"time"

	"github.com/m4n5ter/makabaka/pb/makabaka"
	"github.com/m4n5ter/makabaka/pkg/util"
	"github.com/m4n5ter/makabaka/rpc/makabaka/internal/svc"
	usermodule "github.com/m4n5ter/makabaka/rpc/makabaka/internal/svc/module/user"

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
func (l *RefreshTokenLogic) RefreshToken(in *makabaka.RefreshTokenRequest) (*makabaka.RefreshTokenResponse, error) {
	if in.RefreshToken == "" {
		return nil, usermodule.ErrInvalidInput
	}

	aes := util.NewAES[map[string]int64](l.svcCtx.Config.MAC.Secret)
	refreshTokenMap, err := aes.Decrypt(in.RefreshToken)
	if err != nil {
		return nil, usermodule.ErrInvalidRefreshToken
	}

	// TODO: refresh token 应该可以被销毁，这里应该检查是否被销毁

	expireat := refreshTokenMap["expireat"]
	if expireat < time.Now().Unix() {
		return nil, usermodule.ErrExpiredRefreshToken
	}

	// TODO: 生成了新的凭证后应该销毁旧的凭证
	macresp, err := l.svcCtx.GenMACResponse(refreshTokenMap["userid"])
	if err != nil {
		return nil, err
	}

	return &makabaka.RefreshTokenResponse{
		MacId:        macresp.ID,
		MacKey:       macresp.Key,
		MacAlgorithm: macresp.Algorithm,
		RefreshToken: macresp.RefreshToken,
	}, nil
}
