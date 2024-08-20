package auth

import (
	"context"
	"time"

	"google.golang.org/grpc/metadata"
)

const (
	DefaultExpiration    = 5 * time.Minute
	MissingMetadata      = "missing metadata"
	InvalidAuthorization = "invalid authorization"
)

const (
	HTTPMethod   = "m-http-method"
	HTTPURL      = "m-http-url"
	HTTPHostname = "m-http-hostname"
	HTTPPort     = "m-http-port"
)

type Authenticator interface {
	Authenticate(ctx context.Context) (metadata.MD, error)
}
