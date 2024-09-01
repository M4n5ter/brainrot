package model

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ CommentModel = (*customCommentModel)(nil)

type (
	// CommentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customCommentModel.
	CommentModel interface {
		commentModel
		// extendedCommentModel is generated by modified model.tpl
		extendedCommentModel
		FindAllByArticleID(ctx context.Context, articleID uint64) ([]*Comment, error)
	}

	customCommentModel struct {
		*defaultCommentModel
	}
)

// NewCommentModel returns a model for the database table.
func NewCommentModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) CommentModel {
	return &customCommentModel{
		defaultCommentModel: newCommentModel(conn, c, opts...),
	}
}

// FindAllByArticleID finds all comments by article id
func (m *customCommentModel) FindAllByArticleID(ctx context.Context, articleID uint64) ([]*Comment, error) {
	if articleID == 0 {
		return nil, ErrNotFound
	}

	query := fmt.Sprintf("SELECT %s FROM %s WHERE `article_id` = ? AND `status` = 1", commentRows, m.table)
	var resp []*Comment
	err := m.QueryRowsNoCacheCtx(ctx, resp, query, articleID)
	if err != nil {
		return nil, err
	}

	return resp, nil
}