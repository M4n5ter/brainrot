package pinglogic

import (
	"context"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Ping is a simple health check
func (l *PingLogic) Ping(in *brainrot.PingRequest) (*brainrot.PingResponse, error) {
	// todo: add your logic here and delete this line

	return &brainrot.PingResponse{}, nil
}

// Generated by modified logic.tpl

// TODO: 设置一个 0~99 的唯一模块编号，以及模块名称
// var moduleNumberPingLogic = merror.MustRegisterErrorModule(0, "PingLogic")

// var ErrExample = svc.DefineError(merror.Common, moduleNumberPingLogic, 10, "脱敏后的信息", "详细信息")
