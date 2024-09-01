package articlelogic

import (
	"context"
	"fmt"
	"strings"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/rpc/brainrot/internal/svc"
	articlemodule "brainrot/rpc/brainrot/internal/svc/module/article"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type RefreshAllArticlesLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshAllArticlesLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshAllArticlesLogic {
	return &RefreshAllArticlesLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Refresh all articles
func (l *RefreshAllArticlesLogic) RefreshAllArticles(in *brainrot.RefreshAllArticlesRequest) (*brainrot.RefreshAllArticlesResponse, error) {
	// TODO: Need MQ to refresh all articles asynchronously.
	var modelarticles []*model.Article
	err := l.svcCtx.ArticleModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		query := fmt.Sprintf("SELECT %s FROM %s WHERE status = 1", strings.Join(builder.RawFieldNames(&model.Article{}), ","), "article")
		return session.QueryRowsCtx(ctx, &modelarticles, query)
	})
	if err != nil || len(modelarticles) == 0 {
		return nil, articlemodule.ErrDBError.Wrap("查询文章失败：%v", err)
	}

	articles := make([]*Article, len(modelarticles))
	for i, modelarticle := range modelarticles {
		articles[i] = &Article{
			ID:       int64(modelarticle.Id),
			Title:    modelarticle.Title,
			Tags:     strings.Split(modelarticle.Tags, ","),
			Author:   modelarticle.Author,
			Poster:   modelarticle.Poster,
			Content:  modelarticle.Content,
			PostAt:   modelarticle.CreatedAt.Unix(),
			EditedAt: modelarticle.UpdatedAt.Unix(),
		}
	}

	_, err = l.svcCtx.Meili.Index("articles").AddDocuments(&articles, "id")
	if err != nil {
		return nil, articlemodule.ErrSystemError.Wrap("Meilisearch 添加文章失败：%v", err)
	}

	return &brainrot.RefreshAllArticlesResponse{}, nil
}
