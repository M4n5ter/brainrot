package articlelogic

import (
	"context"
	"strconv"
	"strings"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/rpc/brainrot/internal/svc"
	articlemodule "brainrot/rpc/brainrot/internal/svc/module/article"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type PostArticleLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPostArticleLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostArticleLogic {
	return &PostArticleLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Post article
func (l *PostArticleLogic) PostArticle(in *brainrot.PostArticleRequest) (*brainrot.PostArticleResponse, error) {
	// TODO: 发布文章应该同时将文章内容存储到数据库和 Meilisearch 中

	ids := metadata.ValueFromIncomingContext(l.ctx, "userid")
	if ids == nil {
		return nil, articlemodule.ErrSystemError.Wrap("元数据中不存在 userid")
	}

	useridstr := ids[0]
	userid, err := strconv.Atoi(useridstr)
	if err != nil {
		return nil, articlemodule.ErrAIError
	}

	modelarticle := &model.Article{}
	err = copier.Copy(modelarticle, in)
	if err != nil {
		return nil, articlemodule.ErrCopierCopy.Wrap("%v", err)
	}

	modelarticle.AuthorId = uint64(userid)
	modelarticle.Tags = strings.Join(in.Tags, ",")
	result, err := l.svcCtx.ArticleModel.Insert(l.ctx, modelarticle)
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("插入文章失败：%v", err)
	}

	articleID, err := result.LastInsertId()
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("获取文章 ID 失败：%v", err)
	}

	return &brainrot.PostArticleResponse{ArticleId: uint64(articleID)}, nil
}
