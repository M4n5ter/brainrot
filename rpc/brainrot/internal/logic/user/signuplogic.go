package userlogic

import (
	"context"
	"errors"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/pkg/util/validator"
	"brainrot/rpc/brainrot/internal/svc"
	usermodule "brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/go-sql-driver/mysql"
	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/jsonx"
	"github.com/zeromicro/go-zero/core/logx"
)

type SignUpLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSignUpLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SignUpLogic {
	return &SignUpLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Creates a new user
func (l *SignUpLogic) SignUp(in *brainrot.SignUpRequest) (*brainrot.SignUpResponse, error) {
	if in.Username == "" || in.Email == "" || in.Password == "" {
		return nil, usermodule.ErrLackNecessaryField.Wrap("缺少用户名/邮箱/密码")
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
		var mysqlErr *mysql.MySQLError
		if errors.As(err, &mysqlErr) {
			if mysqlErr.Number == 1062 {
				return nil, usermodule.ErrDBDuplicateUsernameOrEmail
			}
		}
		return nil, usermodule.ErrDBError.Wrap("%v", err)
	}

	nRows, err := ret.RowsAffected()
	if err != nil || nRows == 0 {
		return nil, usermodule.ErrDBError.Wrap("插入数据失败")
	}

	return &brainrot.SignUpResponse{}, nil
}
