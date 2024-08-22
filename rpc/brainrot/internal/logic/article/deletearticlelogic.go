package articlelogic

import (
	"context"
	"strconv"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	articlemodule "brainrot/rpc/brainrot/internal/svc/module/article"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type DeleteArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteArticleLogic {
	return &DeleteArticleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Delete article
func (l *DeleteArticleLogic) DeleteArticle(in *brainrot.DeleteArticleRequest) (*brainrot.DeleteArticleResponse, error) {
	if in.Identifier == nil {
		return nil, articlemodule.ErrLackNecessaryField
	}

	ids := metadata.ValueFromIncomingContext(l.ctx, "userid")
	if ids == nil {
		return nil, articlemodule.ErrSystemError.Wrap("元数据中不存在 userid")
	}

	useridstr := ids[0]
	userid, err := strconv.Atoi(useridstr)
	if err != nil {
		return nil, articlemodule.ErrAIError
	}

	var articleID uint64
	switch identifier := in.Identifier.(type) {
	case *brainrot.DeleteArticleRequest_Id:
		err = l.svcCtx.ArticleModel.LogicDelete(l.ctx, identifier.Id)
		if err != nil {
			return nil, articlemodule.ErrDBError.Wrap("删除文章失败：%v", err)
		}

		articleID = identifier.Id
	case *brainrot.DeleteArticleRequest_Title:
		articleID, err = l.svcCtx.ArticleModel.LogicDeleteByAuthorIDTitle(l.ctx, uint64(userid), identifier.Title)
		if err != nil {
			return nil, articlemodule.ErrDBError.Wrap("删除文章失败：%v", err)
		}
	}

	// TODO: 可以引入 MQ 来处理 meilisearch 未能删除成功的文章
	_, err = l.svcCtx.Meili.Index("articles").DeleteDocument(strconv.Itoa(int(articleID)))
	if err != nil {
		return nil, articlemodule.ErrSystemError.Wrap("从 meilisearch 删除文章失败：%v", err)
	}

	return &brainrot.DeleteArticleResponse{}, nil
}
