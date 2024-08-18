package userlogic

import (
	"context"
	"fmt"

	"github.com/jinzhu/copier"
	"github.com/m4n5ter/brainrot/model"
	"github.com/m4n5ter/brainrot/pb/brainrot"
	"github.com/m4n5ter/brainrot/pkg/util/validator"
	"github.com/m4n5ter/brainrot/rpc/brainrot/internal/svc"
	usermodule "github.com/m4n5ter/brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
)

type SighUpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSighUpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SighUpLogic {
	return &SighUpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Creates a new user
func (l *SighUpLogic) SighUp(in *brainrot.SighUpRequest) (*brainrot.SighUpResponse, error) {
	if in.Username == "" || in.Email == "" || in.Password == "" {
		return nil, fmt.Errorf("缺少用户名/邮箱/密码: %w", usermodule.ErrLackNecessaryField)
	}

	if !validator.IsEmail(in.Email) {
		return nil, usermodule.ErrInvalidInput.Wrap("邮箱格式不合法")
	}

	if in.ProfileInfo != "" {
		// TODO: 恶意构造巨大的 ProfileInfo 会导致问题。也许应该在 api 网关处理
		var pi map[string]any
		if err := jsonx.UnmarshalFromString(in.ProfileInfo, pi); err != nil {
			return nil, usermodule.ErrInvalidInput.Wrap("ProfileInfo 不是合法 JSON 字符串")
		}
		if n := len(pi); n > 10 {
			return nil, usermodule.ErrInvalidInput.Wrap("输入的 ProfileInfo 字段数量为: %d, 超过 10", n)
		}
	}

	usermodel := &model.User{}
	err := copier.Copy(usermodel, in)
	if err != nil {
		return nil, usermodule.ErrCopierCopy.Wrap("%v", err)
	}

	usermodel.ProfileInfo = "{}"

	ret, err := l.svcCtx.UserModel.Insert(l.ctx, usermodel)
	if err != nil {
		return nil, usermodule.ErrDBError.Wrap("%v", err)
	}

	nRows, err := ret.RowsAffected()
	if err != nil || nRows == 0 {
		return nil, usermodule.ErrDBError.Wrap("插入数据失败")
	}

	id, err := ret.LastInsertId()
	if err != nil {
		return nil, usermodule.ErrDBError.Wrap("获取插入数据的 ID 失败")
	}

	return &brainrot.SighUpResponse{UserId: id}, nil
}
