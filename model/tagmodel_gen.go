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
	tagFieldNames          = builder.RawFieldNames(&Tag{})
	tagRows                = strings.Join(tagFieldNames, ",")
	tagRowsExpectAutoSet   = strings.Join(stringx.Remove(tagFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`status`", "`update_at`", "`update_time`", "`updated_at`"), ",")
	tagRowsWithPlaceHolder = strings.Join(stringx.Remove(tagFieldNames, "`id`", "`create_at`", "`create_time`", "`created_at`", "`status`", "`update_at`", "`update_time`", "`updated_at`"), "=?,") + "=?"

	cacheTagIdPrefix   = "cache:tag:id:"
	cacheTagNamePrefix = "cache:tag:name:"
)

type (
	tagModel interface {
		Insert(ctx context.Context, data *Tag) (sql.Result, error)
		FindOne(ctx context.Context, id uint64) (*Tag, error)
		FindOneByName(ctx context.Context, name string) (*Tag, error)
		Update(ctx context.Context, data *Tag) error
		Delete(ctx context.Context, id uint64) error
	}

	defaultTagModel struct {
		sqlc.CachedConn
		table string
	}

	Tag struct {
		Id        uint64    `db:"id"`
		Name      string    `db:"name"`
		CreatedAt time.Time `db:"created_at"`
	}
)

func newTagModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) *defaultTagModel {
	return &defaultTagModel{
		CachedConn: sqlc.NewConn(conn, c, opts...),
		table:      "`tag`",
	}
}

func (m *defaultTagModel) Delete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	tagIdKey := fmt.Sprintf("%s%v", cacheTagIdPrefix, id)
	tagNameKey := fmt.Sprintf("%s%v", cacheTagNamePrefix, data.Name)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("delete from %s where `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, tagIdKey, tagNameKey)
	return err
}

func (m *defaultTagModel) FindOne(ctx context.Context, id uint64) (*Tag, error) {
	tagIdKey := fmt.Sprintf("%s%v", cacheTagIdPrefix, id)
	var resp Tag
	err := m.QueryRowCtx(ctx, &resp, tagIdKey, func(ctx context.Context, conn sqlx.SqlConn, v any) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? AND limit 1", tagRows, m.table)
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

func (m *defaultTagModel) FindOneByName(ctx context.Context, name string) (*Tag, error) {
	tagNameKey := fmt.Sprintf("%s%v", cacheTagNamePrefix, name)
	var resp Tag
	err := m.QueryRowIndexCtx(ctx, &resp, tagNameKey, m.formatPrimary, func(ctx context.Context, conn sqlx.SqlConn, v any) (i any, e error) {
		query := fmt.Sprintf("select %s from %s where `name` = ? limit 1", tagRows, m.table)
		if err := conn.QueryRowCtx(ctx, &resp, query, name); err != nil {
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

func (m *defaultTagModel) Insert(ctx context.Context, data *Tag) (sql.Result, error) {
	tagIdKey := fmt.Sprintf("%s%v", cacheTagIdPrefix, data.Id)
	tagNameKey := fmt.Sprintf("%s%v", cacheTagNamePrefix, data.Name)
	ret, err := m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("insert into %s (%s) values (?)", m.table, tagRowsExpectAutoSet)
		return conn.ExecCtx(ctx, query, data.Name)
	}, tagIdKey, tagNameKey)
	return ret, err
}

func (m *defaultTagModel) Update(ctx context.Context, newData *Tag) error {
	data, err := m.FindOne(ctx, newData.Id)
	if err != nil {
		return err
	}

	tagIdKey := fmt.Sprintf("%s%v", cacheTagIdPrefix, data.Id)
	tagNameKey := fmt.Sprintf("%s%v", cacheTagNamePrefix, data.Name)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, tagRowsWithPlaceHolder)
		return conn.ExecCtx(ctx, query, newData.Name, newData.Id)
	}, tagIdKey, tagNameKey)
	return err
}

func (m *defaultTagModel) formatPrimary(primary any) string {
	return fmt.Sprintf("%s%v", cacheTagIdPrefix, primary)
}

func (m *defaultTagModel) queryPrimary(ctx context.Context, conn sqlx.SqlConn, v, primary any) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", tagRows, m.table)
	return conn.QueryRowCtx(ctx, v, query, primary)
}

func (m *defaultTagModel) tableName() string {
	return m.table
}

// Generated by modified customized.tpl

type extendedTagModel interface {
	Trans(ctx context.Context, fn func(context context.Context, session sqlx.Session) error) error
	LogicDelete(ctx context.Context, id uint64) error
	FindPageListByIdDESC(ctx context.Context, preMinID, pageSize uint64) ([]*Tag, error)
	FindPageListByIdASC(ctx context.Context, preMaxID, pageSize uint64) ([]*Tag, error)
}

func (m *defaultTagModel) LogicDelete(ctx context.Context, id uint64) error {
	data, err := m.FindOne(ctx, id)
	if err != nil {
		return err
	}

	tagIdKey := fmt.Sprintf("%s%v", cacheTagIdPrefix, id)
	tagNameKey := fmt.Sprintf("%s%v", cacheTagNamePrefix, data.Name)
	_, err = m.ExecCtx(ctx, func(ctx context.Context, conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("UPDATE %s SET `status` = 0 WHERE `id` = ?", m.table)
		return conn.ExecCtx(ctx, query, id)
	}, tagIdKey, tagNameKey)
	return err
}

func (m *defaultTagModel) FindPageListByIdDESC(ctx context.Context, preMinID, pageSize uint64) ([]*Tag, error) {
	args := []any{}
	where := " "

	if preMinID > 0 {
		where = " WHERE `id` < ? and `status` = 0"
		args = append(args, preMinID)
	}

	query := fmt.Sprintf("SELECT %s FROM %s%sORDER BY `id` DESC LIMIT ?", tagRows, m.table, where)
	args = append(args, pageSize)

	var resp []*Tag
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

func (m *defaultTagModel) FindPageListByIdASC(ctx context.Context, preMaxID, pageSize uint64) ([]*Tag, error) {
	args := []any{}
	where := " "

	if preMaxID > 0 {
		where = " WHERE `id` > ? and `status` = 0"
		args = append(args, preMaxID)
	}

	query := fmt.Sprintf("SELECT %s FROM %s%sORDER BY `id` ASC LIMIT ?", tagRows, m.table, where)
	args = append(args, pageSize)

	var resp []*Tag
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

func (m *defaultTagModel) Trans(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return m.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return fn(ctx, session)
	})
}
