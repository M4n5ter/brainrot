// Code generated by goctl. DO NOT EDIT.

package model

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/stores/builder"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlc"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stringx"
)

var (
	articleFieldNames          = builder.RawFieldNames(&Article{})
	articleRows                = strings.Join(articleFieldNames, ",")
	articleRowsExpectAutoSet   = strings.Join(stringx.Remove(articleFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`status`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	articleRowsWithPlaceHolder = strings.Join(stringx.Remove(articleFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`status`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"

	cacheArticleIdPrefix            = "cache:article:id:"
	cacheArticleAuthorIdTitlePrefix = "cache:article:authorId:title:"
)

type (
	articleModel interface {
		Insert(ctx context.Context, data *Article) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*Article, error)
		FindOneByAuthorIdTitle(ctx context.Context, authorId uint64, title string) (*Article, error)
		Update(ctx context.Context, data *Article) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultArticleModel struct {
		sqlc.CachedConn
		table string
	}

	Article struct {
		Id        uint64    `db:"id"`
		AuthorId  uint64    `db:"author_id"`
		Author    string    `db:"author"`
		Title     string    `db:"title"`
		Content   string    `db:"content"`
		Tags      string    `db:"tags"`
		Poster    string    `db:"poster"`
		Status    int64     `db:"status"` // 0: active, 1: deleted
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}
)

func newArticleModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultArticleModel {
	return &defaultArticleModel{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`article`",
	}
}

func (m *defaultArticleModel) Delete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	articleAuthorIdTitleKey := fmt.Sprintf("%s%v:%v", cacheArticleAuthorIdTitlePrefix, data.AuthorId, data.Title)
	articleIdKey := fmt.Sprintf("%s%v", cacheArticleIdPrefix, id)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, articleAuthorIdTitleKey, articleIdKey)
	return err
}

func (m *defaultArticleModel) FindOne(ctx context.Context, id uint64) (*Article, error) {
	articleIdKey := fmt.Sprintf("%s%v", cacheArticleIdPrefix, id)
	var resp Article
	err := m.QueryRowCtx(ctx, &resp, articleIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? AND `status` = 1 limit 1", articleRows, m.table)
		return conn.QueryRowCtx(ctx, v, query, id)
	})
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultArticleModel) FindOneByAuthorIdTitle(ctx context.Context, authorId uint64, title string) (*Article, error) {
	articleAuthorIdTitleKey := fmt.Sprintf("%s%v:%v", cacheArticleAuthorIdTitlePrefix, authorId, title)
	var resp Article
	err := m.QueryRowIndexCtx(ctx, &resp, articleAuthorIdTitleKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `author_id` = ? and `title` = ? AND `status` = 1 limit 1", articleRows, m.table)
		if err := conn.QueryRowCtx(ctx, &resp, query, authorId, title); err != nil {
			return nil, err
		}
		return resp.Id, nil
	}, m.queryPrimary)
	switch err {
	case nil:
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultArticleModel) Insert(ctx context.Context, data *Article) (sql.Result, error) {
	articleAuthorIdTitleKey := fmt.Sprintf("%s%v:%v", cacheArticleAuthorIdTitlePrefix, data.AuthorId, data.Title)
	articleIdKey := fmt.Sprintf("%s%v", cacheArticleIdPrefix, data.Id)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?, ?, ?, ?, ?, ?)", m.table, articleRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.AuthorId, data.Author, data.Title, data.Content, data.Tags, data.Poster)
	}, articleAuthorIdTitleKey, articleIdKey)
	return ret, err
}

func (m *defaultArticleModel) Update(ctx context.Context, newData *Article) error {
	data, err := m.FindOne(ctx, newData.Id)
	if err != nil {
		return err
	}

	articleAuthorIdTitleKey := fmt.Sprintf("%s%v:%v", cacheArticleAuthorIdTitlePrefix, data.AuthorId, data.Title)
	articleIdKey := fmt.Sprintf("%s%v", cacheArticleIdPrefix, data.Id)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, articleRowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, newData.AuthorId, newData.Author, newData.Title, newData.Content, newData.Tags, newData.Poster, newData.Id)
	}, articleAuthorIdTitleKey, articleIdKey)
	return err
}

func (m *defaultArticleModel) formatPrimary(primary any) string {
	return fmt.Sprintf("%s%v", cacheArticleIdPrefix, primary)
}

func (m *defaultArticleModel) queryPrimary(ctx context.Context, conn sqlx.SqlConn, v, primary any) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? AND `status` = 1 limit 1", articleRows, m.table)
	return conn.QueryRowCtx(ctx, v, query, primary)
}

func (m *defaultArticleModel) tableName() string {
	return m.table
}

// Generated by modified customized.tpl

type extendedArticleModel interface {
	Trans(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error
	LogicDelete(ctx context.Context, id uint64) error
	FindPageListByIdDESC(ctx context.Context, preMinID, pageSize uint64) ([]*Article, error)
	FindPageListByIdASC(ctx context.Context, preMaxID, pageSize uint64) ([]*Article, error)
}

func (m *defaultArticleModel) LogicDelete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	articleAuthorIdTitleKey := fmt.Sprintf("%s%v:%v", cacheArticleAuthorIdTitlePrefix, data.AuthorId, data.Title)
	articleIdKey := fmt.Sprintf("%s%v", cacheArticleIdPrefix, id)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("UPDATE %s SET `status` = 0 WHERE `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, articleAuthorIdTitleKey, articleIdKey)
	return err
}

func (m *defaultArticleModel) FindPageListByIdDESC(ctx context.Context, preMinID, pageSize uint64) ([]*Article, error) {
	args := []any{}
	where := " "

	if preMinID > 0 {
		where = " WHERE `id` < ? and `status` = 0"
		args = append(args, preMinID)
	}

	query := fmt.Sprintf("SELECT %s FROM %s%sORDER BY `id` DESC LIMIT ?", articleRows, m.table, where)
	args = append(args, pageSize)

	var resp []*Article
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultArticleModel) FindPageListByIdASC(ctx context.Context, preMaxID, pageSize uint64) ([]*Article, error) {
	args := []any{}
	where := " "

	if preMaxID > 0 {
		where = " WHERE `id` > ? and `status` = 0"
		args = append(args, preMaxID)
	}

	query := fmt.Sprintf("SELECT %s FROM %s%sORDER BY `id` ASC LIMIT ?", articleRows, m.table, where)
	args = append(args, pageSize)

	var resp []*Article
	err := m.QueryRowsNoCacheCtx(ctx, &resp, query, args...)
	switch err {
	case nil:
		return resp, nil
	case sqlx.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultArticleModel) Trans(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return fn(ctx, session)
	})
}
