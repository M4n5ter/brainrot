package gatewayoption

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	_ "unsafe" // required for go:linkname

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
)

func ErrorHandler(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
	// return Internal when Marshal failed
	const fallback = `{"code": 13, "message": "failed to marshal error message"}`

	var customStatus *runtime.HTTPStatusError
	if errors.As(err, &customStatus) {
		err = customStatus.Err
	}

	s := status.Convert(err)
	pb := s.Proto()

	w.Header().Del("Trailer")
	w.Header().Del("Transfer-Encoding")

	contentType := marshaler.ContentType(pb)
	w.Header().Set("Content-Type", contentType)

	if s.Code() == codes.Unauthenticated {
		w.Header().Set("WWW-Authenticate", s.Message())
	}

	buf, merr := marshaler.Marshal(pb)
	if merr != nil {
		grpclog.Errorf("Failed to marshal error message %q: %v", s, merr)
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			grpclog.Errorf("Failed to write response: %v", err)
		}
		return
	}

	md, ok := runtime.ServerMetadataFromContext(ctx)
	if !ok {
		grpclog.Error("Failed to extract ServerMetadata from context")
	}

	handleForwardResponseServerMetadata(w, mux, md)

	// RFC 7230 https://tools.ietf.org/html/rfc7230#section-4.1.2
	// Unless the request includes a TE header field indicating "trailers"
	// is acceptable, as described in Section 4.3, a server SHOULD NOT
	// generate trailer fields that it believes are necessary for the user
	// agent to receive.
	doForwardTrailers := requestAcceptsTrailers(r)

	if doForwardTrailers {
		handleForwardResponseTrailerHeader(w, mux, md)
		w.Header().Set("Transfer-Encoding", "chunked")
	}

	st := runtime.HTTPStatusFromCode(s.Code())
	// set http status code
	if customStatus != nil {
		st = customStatus.HTTPStatus
		w.WriteHeader(st)
	} else if vals := md.HeaderMD.Get("x-http-code"); len(vals) > 0 {
		code, err := strconv.Atoi(vals[0])
		if err != nil {
			return
		}
		// delete the headers to not expose any grpc-metadata in http response
		delete(md.HeaderMD, "x-http-code")
		delete(w.Header(), "Grpc-Metadata-X-Http-Code")
		w.WriteHeader(code)
	} else {
		w.WriteHeader(st)
	}

	if _, err := w.Write(buf); err != nil {
		grpclog.Errorf("Failed to write response: %v", err)
	}

	if doForwardTrailers {
		handleForwardResponseTrailer(w, mux, md)
	}
}

//go:linkname requestAcceptsTrailers github.com/grpc-ecosystem/grpc-gateway/v2/runtime.requestAcceptsTrailers
func requestAcceptsTrailers(req *http.Request) bool

//go:linkname handleForwardResponseServerMetadata github.com/grpc-ecosystem/grpc-gateway/v2/runtime.handleForwardResponseServerMetadata
func handleForwardResponseServerMetadata(w http.ResponseWriter, mux *runtime.ServeMux, md runtime.ServerMetadata)

//go:linkname handleForwardResponseTrailerHeader github.com/grpc-ecosystem/grpc-gateway/v2/runtime.handleForwardResponseTrailerHeader
func handleForwardResponseTrailerHeader(w http.ResponseWriter, mux *runtime.ServeMux, md runtime.ServerMetadata)

//go:linkname handleForwardResponseTrailer github.com/grpc-ecosystem/grpc-gateway/v2/runtime.handleForwardResponseTrailer
func handleForwardResponseTrailer(w http.ResponseWriter, mux *runtime.ServeMux, md runtime.ServerMetadata)
