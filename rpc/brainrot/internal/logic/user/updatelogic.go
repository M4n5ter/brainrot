package userlogic

import (
	"context"
	"errors"
	"strconv"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/rpc/brainrot/internal/svc"
	usermodule "brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/jinzhu/copier"
	"google.golang.org/grpc/metadata"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateLogic {
	return &UpdateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update user
func (l *UpdateLogic) Update(in *brainrot.UpdateUserRequest) (*brainrot.UpdateUserResponse, error) {
	ids := metadata.ValueFromIncomingContext(l.ctx, "userid")
	if ids == nil {
		return nil, usermodule.ErrSystemError.Wrap("元数据中不存在 userid")
	}

	useridstr := ids[0]
	userid, err := strconv.Atoi(useridstr)
	if err != nil {
		return nil, usermodule.ErrAIError
	}

	modeluser, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userid))
	if errors.Is(err, model.ErrNotFound) {
		return nil, usermodule.ErrDBError.Wrap("用户ID %d 不存在", userid)
	}
	if err != nil {
		return nil, usermodule.ErrDBError
	}

	err = copier.CopyWithOption(modeluser, in, copier.Option{IgnoreEmpty: true})
	if err != nil {
		return nil, usermodule.ErrCopierCopy.Wrap("%v", err)
	}

	err = l.svcCtx.UserModel.Update(l.ctx, modeluser)
	if err != nil {
		return nil, usermodule.ErrDBError.Wrap("更新用户失败：%v", err)
	}

	return &brainrot.UpdateUserResponse{}, nil
}
