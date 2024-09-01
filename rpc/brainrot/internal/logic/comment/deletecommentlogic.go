package commentlogic

import (
	"context"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	commentmodule "brainrot/rpc/brainrot/internal/svc/module/comment"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Delete comment
func (l *DeleteCommentLogic) DeleteComment(in *brainrot.DeleteCommentRequest) (*brainrot.DeleteCommentResponse, error) {
	if in.CommentId == 0 {
		return nil, commentmodule.ErrLackNecessaryField
	}

	_, err := l.svcCtx.CommentModel.FindOne(l.ctx, in.CommentId)
	if err != nil {
		return nil, commentmodule.ErrInvalidInput.Wrap("评论 %d 不存在", in.CommentId)
	}

	err = l.svcCtx.CommentModel.LogicDelete(l.ctx, in.CommentId)
	if err != nil {
		return nil, commentmodule.ErrDBError.Wrap("逻辑删除评论 %d 失败", in.CommentId)
	}

	return &brainrot.DeleteCommentResponse{}, nil
}
