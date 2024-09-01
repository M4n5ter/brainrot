package model

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

var _ TagModel = (*customTagModel)(nil)

type (
	// TagModel is an interface to be customized, add more methods here,
	// and implement the added methods in customTagModel.
	TagModel interface {
		tagModel
		// extendedTagModel is generated by modified model.tpl
		extendedTagModel
		BulkInsertTags(sqlConn sqlx.SqlConn, tags []string) error
	}

	customTagModel struct {
		*defaultTagModel
	}
)

// NewTagModel returns a model for the database table.
func NewTagModel(conn sqlx.SqlConn, c cache.CacheConf, opts ...cache.Option) TagModel {
	return &customTagModel{
		defaultTagModel: newTagModel(conn, c, opts...),
	}
}

// BulkInsertTags inserts multiple records into the table.
func (m *customTagModel) BulkInsertTags(sqlConn sqlx.SqlConn, tags []string) error {
	if len(tags) == 0 {
		return nil
	}

	blk, err := sqlx.NewBulkInserter(sqlConn, "INSERT INTO `tag` (`name`) VALUES (?)")
	if err != nil {
		return err
	}

	for _, tag := range tags {
		_ = blk.Insert(tag)
	}

	return nil
}
