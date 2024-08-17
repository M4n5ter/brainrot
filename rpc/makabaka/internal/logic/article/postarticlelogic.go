package articlelogic

import (
	"context"

	"github.com/m4n5ter/makabaka/pb/makabaka"
	"github.com/m4n5ter/makabaka/rpc/makabaka/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
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
func (l *PostArticleLogic) PostArticle(in *makabaka.PostArticleRequest) (*makabaka.PostArticleResponse, error) {
	// todo: add your logic here and delete this line

	return &makabaka.PostArticleResponse{}, nil
}

// Generated by modified logic.tpl

// TODO: 设置一个 0~99 的唯一模块编号，以及模块名称
// var moduleNumberPostArticleLogic = merror.MustRegisterErrorModule(0, "PostArticleLogic")

// var ErrExample = merror.DefineError(merror.Common, moduleNumberPostArticleLogic, 10, "脱敏后的信息", "详细信息")