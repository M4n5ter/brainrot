// Code generated by goctl. DO NOT EDIT.
// Source: makabaka.proto

package server

import (
	"context"

	"github.com/m4n5ter/makabaka/pb/makabaka"
	"github.com/m4n5ter/makabaka/rpc/makabaka/internal/logic/user"
	"github.com/m4n5ter/makabaka/rpc/makabaka/internal/svc"
)

type UserServer struct {
	svcCtx *svc.ServiceContext
	makabaka.UnimplementedUserServer
}

func NewUserServer(svcCtx *svc.ServiceContext) *UserServer {
	return &UserServer{
		svcCtx: svcCtx,
	}
}

// Sigh up
func (s *UserServer) SighUp(ctx context.Context, in *makabaka.SighUpRequest) (*makabaka.SighUpResponse, error) {
	l := userlogic.NewSighUpLogic(ctx, s.svcCtx)
	return l.SighUp(in)
}

// Sigh in
func (s *UserServer) SighIn(ctx context.Context, in *makabaka.SighInRequest) (*makabaka.SighInResponse, error) {
	l := userlogic.NewSighInLogic(ctx, s.svcCtx)
	return l.SighIn(in)
}

// Refresh token
func (s *UserServer) RefreshToken(ctx context.Context, in *makabaka.RefreshTokenRequest) (*makabaka.RefreshTokenResponse, error) {
	l := userlogic.NewRefreshTokenLogic(ctx, s.svcCtx)
	return l.RefreshToken(in)
}
