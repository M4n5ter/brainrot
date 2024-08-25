package userlogic

import (
	"context"
	"strconv"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	usermodule "brainrot/rpc/brainrot/internal/svc/module/user"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type GetCurrentUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCurrentUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCurrentUserInfoLogic {
	return &GetCurrentUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get current user info
func (l *GetCurrentUserInfoLogic) GetCurrentUserInfo(in *brainrot.GetCurrentUserInfoRequest) (*brainrot.GetCurrentUserInfoResponse, error) {
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
	if err != nil {
		return nil, usermodule.ErrSystemError.Wrap("查询用户失败: %v", err)
	}

	resp := &brainrot.GetCurrentUserInfoResponse{}
	err = copier.Copy(resp, modeluser)
	if err != nil {
		return nil, usermodule.ErrCopierCopy.Wrap("%v", err)
	}

	return resp, nil
}
