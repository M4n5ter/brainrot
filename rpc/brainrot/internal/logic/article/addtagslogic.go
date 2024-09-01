package articlelogic

import (
	"context"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	articlemodule "brainrot/rpc/brainrot/internal/svc/module/article"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type AddTagsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddTagsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddTagsLogic {
	return &AddTagsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Add tags
func (l *AddTagsLogic) AddTags(in *brainrot.AddTagsRequest) (*brainrot.AddTagsResponse, error) {
	if len(in.Tags) == 0 {
		return nil, articlemodule.ErrLackNecessaryField
	}

	sqlConn := sqlx.NewMysql(l.svcCtx.Config.MysqlDataSource)
	err := l.svcCtx.TagModel.BulkInsertTags(sqlConn, in.Tags)
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("批量插入标签失败：%v", err)
	}

	return &brainrot.AddTagsResponse{}, nil
}
