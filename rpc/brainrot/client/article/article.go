// Code generated by goctl. DO NOT EDIT.
// Source: brainrot.proto

package article

import (
	"context"

	"brainrot/gen/pb/brainrot"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	AddTagsRequest                       = brainrot.AddTagsRequest
	AddTagsResponse                      = brainrot.AddTagsResponse
	DeleteArticleRequest                 = brainrot.DeleteArticleRequest
	DeleteArticleResponse                = brainrot.DeleteArticleResponse
	DeleteCommentRequest                 = brainrot.DeleteCommentRequest
	DeleteCommentResponse                = brainrot.DeleteCommentResponse
	DeleteTagRequest                     = brainrot.DeleteTagRequest
	DeleteTagResponse                    = brainrot.DeleteTagResponse
	EditCommentRequest                   = brainrot.EditCommentRequest
	EditCommentResponse                  = brainrot.EditCommentResponse
	Error                                = brainrot.Error
	GetCommentsByArticleRequest          = brainrot.GetCommentsByArticleRequest
	GetCommentsByArticleResponse         = brainrot.GetCommentsByArticleResponse
	GetCommentsByArticleResponse_Comment = brainrot.GetCommentsByArticleResponse_Comment
	GetCurrentUserInfoRequest            = brainrot.GetCurrentUserInfoRequest
	GetCurrentUserInfoResponse           = brainrot.GetCurrentUserInfoResponse
	GetPresignedURLRequest               = brainrot.GetPresignedURLRequest
	GetPresignedURLResponse              = brainrot.GetPresignedURLResponse
	MacFields                            = brainrot.MacFields
	PingRequest                          = brainrot.PingRequest
	PingResponse                         = brainrot.PingResponse
	PostArticleRequest                   = brainrot.PostArticleRequest
	PostArticleResponse                  = brainrot.PostArticleResponse
	PostCommentRequest                   = brainrot.PostCommentRequest
	PostCommentResponse                  = brainrot.PostCommentResponse
	RefreshAllArticlesRequest            = brainrot.RefreshAllArticlesRequest
	RefreshAllArticlesResponse           = brainrot.RefreshAllArticlesResponse
	RefreshTokenRequest                  = brainrot.RefreshTokenRequest
	RefreshTokenResponse                 = brainrot.RefreshTokenResponse
	SearchUsersRequest                   = brainrot.SearchUsersRequest
	SearchUsersResponse                  = brainrot.SearchUsersResponse
	SearchUsersResponse_User             = brainrot.SearchUsersResponse_User
	SignInRequest                        = brainrot.SignInRequest
	SignInResponse                       = brainrot.SignInResponse
	SignUpRequest                        = brainrot.SignUpRequest
	SignUpResponse                       = brainrot.SignUpResponse
	UpdateCommentUsefulnessRequest       = brainrot.UpdateCommentUsefulnessRequest
	UpdateCommentUsefulnessResponse      = brainrot.UpdateCommentUsefulnessResponse
	UpdateUserRequest                    = brainrot.UpdateUserRequest
	UpdateUserResponse                   = brainrot.UpdateUserResponse

	Article interface {
		// Post article
		PostArticle(ctx context.Context, in *PostArticleRequest, opts ...grpc.CallOption) (*PostArticleResponse, error)
		// Delete article
		DeleteArticle(ctx context.Context, in *DeleteArticleRequest, opts ...grpc.CallOption) (*DeleteArticleResponse, error)
		// Add tags
		AddTags(ctx context.Context, in *AddTagsRequest, opts ...grpc.CallOption) (*AddTagsResponse, error)
		// Delete tags
		DeleteTag(ctx context.Context, in *DeleteTagRequest, opts ...grpc.CallOption) (*DeleteTagResponse, error)
		// Refresh all articles
		RefreshAllArticles(ctx context.Context, in *RefreshAllArticlesRequest, opts ...grpc.CallOption) (*RefreshAllArticlesResponse, error)
	}

	defaultArticle struct {
		cli zrpc.Client
	}
)

func NewArticle(cli zrpc.Client) Article {
	return &defaultArticle{
		cli: cli,
	}
}

// Post article
func (m *defaultArticle) PostArticle(ctx context.Context, in *PostArticleRequest, opts ...grpc.CallOption) (*PostArticleResponse, error) {
	client := brainrot.NewArticleClient(m.cli.Conn())
	return client.PostArticle(ctx, in, opts...)
}

// Delete article
func (m *defaultArticle) DeleteArticle(ctx context.Context, in *DeleteArticleRequest, opts ...grpc.CallOption) (*DeleteArticleResponse, error) {
	client := brainrot.NewArticleClient(m.cli.Conn())
	return client.DeleteArticle(ctx, in, opts...)
}

// Add tags
func (m *defaultArticle) AddTags(ctx context.Context, in *AddTagsRequest, opts ...grpc.CallOption) (*AddTagsResponse, error) {
	client := brainrot.NewArticleClient(m.cli.Conn())
	return client.AddTags(ctx, in, opts...)
}

// Delete tags
func (m *defaultArticle) DeleteTag(ctx context.Context, in *DeleteTagRequest, opts ...grpc.CallOption) (*DeleteTagResponse, error) {
	client := brainrot.NewArticleClient(m.cli.Conn())
	return client.DeleteTag(ctx, in, opts...)
}

// Refresh all articles
func (m *defaultArticle) RefreshAllArticles(ctx context.Context, in *RefreshAllArticlesRequest, opts ...grpc.CallOption) (*RefreshAllArticlesResponse, error) {
	client := brainrot.NewArticleClient(m.cli.Conn())
	return client.RefreshAllArticles(ctx, in, opts...)
}
