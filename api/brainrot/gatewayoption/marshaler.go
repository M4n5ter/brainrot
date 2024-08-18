package gatewayoption

import (
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
)

type ResponseWrapper struct {
	runtime.JSONPb
}

func NewResponseWrapper() (string, *ResponseWrapper) {
	return runtime.MIMEWildcard, &ResponseWrapper{}
}

func (rw *ResponseWrapper) Marshal(data any) ([]byte, error) {
	var resp any

	switch v := data.(type) {
	case *status.Status:
		message := v.GetMessage()

		switch codes.Code(v.Code) {
		case codes.Internal:
			message = "服务器正忙，请稍后再试"
		}

		resp = map[string]any{
			"code":    v.GetCode(),
			"message": message,
		}
	default:
		resp = map[string]any{
			"code":    0,
			"message": "ok",
			"data":    data,
		}
	}

	return rw.JSONPb.Marshal(resp)
}
