package userlogic

import (
	"context"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/rpc/brainrot/internal/svc"
	usermodule "brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type SearchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Search users
func (l *SearchLogic) Search(in *brainrot.SearchUsersRequest) (*brainrot.SearchUsersResponse, error) {
	if in.Email == "" && in.Username == "" {
		return nil, usermodule.ErrLackNecessaryField.Wrap("Email or Username is required")
	}

	if in.Email != "" && in.Username != "" {
		modelusers, err := l.svcCtx.UserModel.SearchUsers(l.ctx, in.Email, in.Username)
		if err != nil {
			return nil, usermodule.ErrDBError.Wrap("%v", err)
		}

		users := make([]*brainrot.SearchUsersResponse_User, len(modelusers))
		err = copier.Copy(&users, &modelusers)
		if err != nil {
			return nil, usermodule.ErrCopierCopy.Wrap("%v", err)
		}

		return &brainrot.SearchUsersResponse{Users: users}, nil
	}

	var modeluser *model.User
	var err error

	if in.Email != "" {
		modeluser, err = l.svcCtx.UserModel.FindOneByEmail(l.ctx, in.Email)
	}
	if in.Username != "" {
		modeluser, err = l.svcCtx.UserModel.FindOneByUsername(l.ctx, in.Username)
	}
	if err != nil {
		return nil, usermodule.ErrDBError.Wrap("%v", err)
	}

	users := make([]*brainrot.SearchUsersResponse_User, 1)
	err = copier.Copy(&users, &modeluser)
	if err != nil {
		return nil, usermodule.ErrCopierCopy.Wrap("%v", err)
	}

	return &brainrot.SearchUsersResponse{Users: users}, nil
}
