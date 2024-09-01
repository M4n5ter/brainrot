package commentlogic

import (
	"context"
	"strconv"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	commentmodule "brainrot/rpc/brainrot/internal/svc/module/comment"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type EditCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEditCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EditCommentLogic {
	return &EditCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Edit comment
func (l *EditCommentLogic) EditComment(in *brainrot.EditCommentRequest) (*brainrot.EditCommentResponse, error) {
	if in.CommentId == 0 || in.Content == "" {
		return nil, commentmodule.ErrLackNecessaryField
	}
	modelcomment, err := l.svcCtx.CommentModel.FindOne(l.ctx, in.CommentId)
	if err != nil {
		return nil, commentmodule.ErrInvalidInput.Wrap("评论 %d 不存在", in.CommentId)
	}

	ids := metadata.ValueFromIncomingContext(l.ctx, "userid")
	if ids == nil {
		return nil, commentmodule.ErrSystemError.Wrap("元数据中不存在 userid")
	}

	useridstr := ids[0]
	userid, err := strconv.Atoi(useridstr)
	if err != nil {
		return nil, commentmodule.ErrAIError
	}

	if modelcomment.UserId != uint64(userid) {
		return nil, commentmodule.ErrNoPermission.Wrap("用户 %d 无权限编辑评论 %d", userid, in.CommentId)
	}

	modelcomment.Content = in.Content
	err = l.svcCtx.CommentModel.Update(l.ctx, modelcomment)
	if err != nil {
		return nil, commentmodule.ErrDBError.Wrap("更新评论 %d 失败", in.CommentId)
	}

	return &brainrot.EditCommentResponse{}, nil
}
