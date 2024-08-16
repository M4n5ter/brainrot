package pinglogic

import (
	"context"

	"github.com/m4n5ter/makabaka/pb/makabaka"
	"github.com/m4n5ter/makabaka/rpc/makabaka/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type PingModPingLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPingModPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingModPingLogic {
	return &PingModPingLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Ping is a simple health check
func (l *PingModPingLogic) PingModPing(in *makabaka.PingRequest) (*makabaka.PingResponse, error) {
	// todo: add your logic here and delete this line

	return &makabaka.PingResponse{}, nil
}

// Generated by modified logic.tpl

// TODO: 设置一个 0~99 的唯一模块编号，以及模块名称
// var moduleNumberPingModPingLogic = merror.MustRegisterErrorModule(0, "PingModPingLogic")

// var ErrExample = svc.DefineError(merror.Common, moduleNumberPingModPingLogic, 10, "脱敏后的信息", "详细信息")
