// Code generated by goctl. DO NOT EDIT.
// Source: makabaka.proto

package ping

import (
	"context"

	"github.com/m4n5ter/makabaka/pb/makabaka"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	PingRequest          = makabaka.PingRequest
	PingResponse         = makabaka.PingResponse
	RefreshTokenRequest  = makabaka.RefreshTokenRequest
	RefreshTokenResponse = makabaka.RefreshTokenResponse
	SighInRequest        = makabaka.SighInRequest
	SighInResponse       = makabaka.SighInResponse
	SighUpRequest        = makabaka.SighUpRequest
	SighUpResponse       = makabaka.SighUpResponse

	Ping interface {
		// Ping is a simple health check
		Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error)
	}

	defaultPing struct {
		cli zrpc.Client
	}
)

func NewPing(cli zrpc.Client) Ping {
	return &defaultPing{
		cli: cli,
	}
}

// Ping is a simple health check
func (m *defaultPing) Ping(ctx context.Context, in *PingRequest, opts ...grpc.CallOption) (*PingResponse, error) {
	client := makabaka.NewPingClient(m.cli.Conn())
	return client.Ping(ctx, in, opts...)
}