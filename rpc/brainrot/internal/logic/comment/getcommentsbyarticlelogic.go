package commentlogic

import (
	"context"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	commentmodule "brainrot/rpc/brainrot/internal/svc/module/comment"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentsByArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentsByArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentsByArticleLogic {
	return &GetCommentsByArticleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get comments by article
func (l *GetCommentsByArticleLogic) GetCommentsByArticle(in *brainrot.GetCommentsByArticleRequest) (*brainrot.GetCommentsByArticleResponse, error) {
	if in.ArticleId == 0 {
		return nil, commentmodule.ErrLackNecessaryField
	}

	modelcomments, err := l.svcCtx.CommentModel.FindAllByArticleID(l.ctx, in.ArticleId)
	if err != nil {
		return nil, commentmodule.ErrDBError.Wrap("查找文章 %d 的评论失败", in.ArticleId)
	}

	comments := make([]*brainrot.GetCommentsByArticleResponse_Comment, len(modelcomments))
	err = copier.Copy(&comments, &modelcomments)
	if err != nil {
		return nil, commentmodule.ErrCopierCopy.Wrap("拷贝评论失败：%v", err)
	}

	return &brainrot.GetCommentsByArticleResponse{Comments: comments}, nil
}
