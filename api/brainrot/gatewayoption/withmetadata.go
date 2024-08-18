package gatewayoption

import (
	"context"
	"net/http"
	"strings"

	"google.golang.org/grpc/metadata"
)

func WithMetadata(ctx context.Context, req *http.Request) metadata.MD {
	httpMethod := req.Method
	httpURL := req.URL.String()
	host := strings.Split(req.Host, ":")
	hostName := host[0]
	port := ""
	if len(host) > 1 {
		port = host[1]
	}

	return metadata.Pairs(
		"m-http-method", httpMethod,
		"m-http-url", httpURL,
		"m-http-hostname", hostName,
		"m-http-port", port,
	)
}
