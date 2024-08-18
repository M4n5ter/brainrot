package unaryserverinterceptors

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/m4n5ter/brainrot/pkg/grpcinterceptor/auth"
	"github.com/m4n5ter/brainrot/pkg/merror"
	"github.com/zeromicro/go-zero/core/logx"
)

// 将 pkg/error/error.go 中的 Error 结构体改为 grpc 的 status.Status 结构体。
func ErrorToStatusInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		// 日志中应记录原始详细信息，转化为 grpc.Status 时需要抹除详细信息。
		logx.Error(err)

		err = merror.UnwrapAll(err)

		merr := merror.NewError(0, "")
		if ok := errors.As(err, &merr); ok {
			code := merr.GetCode()
			if strings.HasPrefix(strconv.Itoa(int(code)), "2") {
				err := grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "422"))
				if err != nil {
					logx.Error(err)
				}
			}

			msg := merror.Desensitize(code)
			err = status.Error(codes.Code(code), msg)
		}
	}

	return resp, err
}

func OAuth2MACAuthorizeInterceptor(authenticator *auth.Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		if err := authenticator.Authenticate(ctx); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}
