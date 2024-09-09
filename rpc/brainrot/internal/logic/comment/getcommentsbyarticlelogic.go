package commentlogic

import (
	"context"
	"errors"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/rpc/brainrot/internal/svc"
	commentmodule "brainrot/rpc/brainrot/internal/svc/module/comment"

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
	if err != nil && !errors.Is(err, model.ErrNotFound) {
		return nil, commentmodule.ErrDBError.Wrap("查找文章 %d 的评论失败，错误为：%v", in.ArticleId, err)
	}

	comments := make([]*brainrot.GetCommentsByArticleResponse_Comment, len(modelcomments))
	for i, comment := range modelcomments {
		comments[i] = &brainrot.GetCommentsByArticleResponse_Comment{
			CommentId:    comment.Id,
			Content:      comment.Content,
			Commenter:    comment.Commenter,
			UsefulCount:  comment.UsefulCount,
			UselessCount: comment.UselessCount,
			CreatedAt:    comment.CreatedAt.Local().Unix(),
			UpdatedAt:    comment.UpdatedAt.Local().Unix(),
		}
	}
	if err != nil {
		return nil, commentmodule.ErrCopierCopy.Wrap("拷贝评论失败：%v", err)
	}

	return &brainrot.GetCommentsByArticleResponse{Comments: comments}, nil
}
