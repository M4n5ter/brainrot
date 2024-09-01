package articlelogic

import (
	"context"
	"strings"

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

	currentMaxID, err := l.svcCtx.TagModel.MaxID(l.ctx)
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("获取标签最大 ID 失败：%v", err)
	}

	type tag struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}

	tags := make([]tag, len(in.Tags))

	for i := range in.Tags {
		in.Tags[i] = strings.ToLower(in.Tags[i])
		tags[i] = tag{ID: currentMaxID + int64(i) + 1, Name: in.Tags[i]}
	}

	sqlConn := sqlx.NewMysql(l.svcCtx.Config.MysqlDataSource)
	err = l.svcCtx.TagModel.BulkInsertTags(sqlConn, in.Tags)
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("批量插入标签失败：%v", err)
	}

	_, err = l.svcCtx.Meili.Index("tags").AddDocuments(tags)
	if err != nil {
		return nil, articlemodule.ErrSystemError.Wrap("添加标签到 MeiliSearch 失败：%v", err)
	}

	return &brainrot.AddTagsResponse{}, nil
}
