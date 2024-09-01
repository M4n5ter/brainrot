package articlelogic

import (
	"context"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	articlemodule "brainrot/rpc/brainrot/internal/svc/module/article"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteTagLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteTagLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteTagLogic {
	return &DeleteTagLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Delete tags
func (l *DeleteTagLogic) DeleteTag(in *brainrot.DeleteTagRequest) (*brainrot.DeleteTagResponse, error) {
	if in.Tag == "" {
		return nil, articlemodule.ErrLackNecessaryField
	}

	modeltag, err := l.svcCtx.TagModel.FindOneByName(l.ctx, in.Tag)
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("查询标签失败：%v", err)
	}

	err = l.svcCtx.TagModel.Delete(l.ctx, modeltag.Id)
	if err != nil {
		return nil, articlemodule.ErrDBError.Wrap("删除标签失败：%v", err)
	}

	return &brainrot.DeleteTagResponse{}, nil
}
