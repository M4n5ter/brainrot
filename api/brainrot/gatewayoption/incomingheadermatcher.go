package gatewayoption

import "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

func HeaderMatcher(key string) (string, bool) {
	switch key {
	case "Authorization":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}
