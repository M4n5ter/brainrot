package userlogic

import (
	"context"

	"github.com/jinzhu/copier"
	"github.com/m4n5ter/makabaka/pb/makabaka"
	"github.com/m4n5ter/makabaka/rpc/makabaka/internal/svc"
	usermodule "github.com/m4n5ter/makabaka/rpc/makabaka/internal/svc/module/user"

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
func (l *SighInLogic) SighIn(in *makabaka.SighInRequest) (*makabaka.SighInResponse, error) {
	if in.Email == "" || in.Password == "" {
		return nil, usermodule.ErrLackNecessaryField.Wrap("邮箱或密码为空")
	}

	usermodel, err := l.svcCtx.UserModel.FindOneByEmail(l.ctx, in.Email)
	if err != nil || usermodel == nil {
		return nil, usermodule.ErrDBError.Wrap("根据邮箱查询用户失败，邮箱为：%s", in.Email)
	}

	if usermodel.Password != in.Password {
		return nil, usermodule.ErrInvalidInput.Wrap("密码错误，邮箱为：%s", in.Email)
	}

	macresp, err := l.svcCtx.GenMACResponse(usermodel.Id)
	if err != nil {
		return nil, err
	}

	resp := &makabaka.SighInResponse{}
	err = copier.Copy(resp, usermodel)
	if err != nil {
		return nil, usermodule.ErrCopierCopy.Wrap("拷贝用户信息失败")
	}

	resp.MacId = macresp.ID
	resp.MacKey = macresp.Key
	resp.MacAlgorithm = macresp.Algorithm
	resp.RefreshToken = macresp.RefreshToken
	return resp, nil
}