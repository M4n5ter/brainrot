package articlelogic

import (
	"context"
	"strconv"
	"strings"
	"time"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/rpc/brainrot/internal/svc"
	articlemodule "brainrot/rpc/brainrot/internal/svc/module/article"

	"github.com/jinzhu/copier"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
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
	ids := metadata.ValueFromIncomingContext(l.ctx, "userid")
	if ids == nil {
		return nil, articlemodule.ErrSystemError.Wrap("元数据中不存在 userid")
	}

	useridstr := ids[0]
	userid, err := strconv.Atoi(useridstr)
	if err != nil {
		return nil, articlemodule.ErrAIError
	}

	if len(in.Tags) > 0 {
		err = l.svcCtx.TagModel.Trans(l.ctx, func(context context.Context, session sqlx.Session) error {
			for _, tag := range in.Tags {
				_, err = l.svcCtx.TagModel.FindOneByName(l.ctx, tag)
				if err != nil {
					return articlemodule.ErrInvalidInput.Wrap("用户 %d 尝试使用不存在的标签 %s", userid, tag)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	modelarticle := &model.Article{}
	err = copier.Copy(modelarticle, in)
	if err != nil {
		return nil, articlemodule.ErrCopierCopy.Wrap("%v", err)
	}

	author, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userid))
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("获取作者信息失败：%v", err)
	}

	modelarticle.AuthorId = uint64(userid)
	modelarticle.Author = author.Username
	modelarticle.Tags = strings.Join(in.Tags, ",")
	result, err := l.svcCtx.ArticleModel.Insert(l.ctx, modelarticle)
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("插入文章失败：%v", err)
	}

	// 发布文章和更新文章的时间可以允许有一些误差，所以在这里直接生成当前时间，而不是再次查询数据库
	now := time.Now()
	modelarticle.CreatedAt = now
	modelarticle.UpdatedAt = now

	articleID, err := result.LastInsertId()
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("获取文章 ID 失败：%v", err)
	}

	// 将文章添加到 meilisearch
	article := &Article{
		ID:       articleID,
		Title:    modelarticle.Title,
		Tags:     in.Tags,
		Author:   author.Username,
		Poster:   modelarticle.Poster,
		Content:  modelarticle.Content,
		PostAt:   modelarticle.CreatedAt.Unix(),
		EditedAt: modelarticle.UpdatedAt.Unix(),
	}

	// TODO: 可以引入 MQ 来处理 meilisearch 未能发布成功的文章
	_, err = l.svcCtx.Meili.Index("articles").AddDocuments(article, "id")
	if err != nil {
		return nil, articlemodule.ErrSystemError.Wrap("Meilisearch 添加文章失败：%v", err)
	}

	return &brainrot.PostArticleResponse{ArticleId: uint64(articleID)}, nil
}

type Article struct {
	ID       int64    `json:"id"`
	Title    string   `json:"title"`
	Tags     []string `json:"tags"`
	Author   string   `json:"author"`
	Poster   string   `json:"poster"`
	Content  string   `json:"content"`
	PostAt   int64    `json:"post_at"`
	EditedAt int64    `json:"edited_at"`
}
