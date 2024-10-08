// Code generated by goctl. DO NOT EDIT.
// Source: brainrot.proto

package server

import (
	"context"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/logic/article"
	"brainrot/rpc/brainrot/internal/svc"
)

type ArticleServer struct {
	svcCtx *svc.ServiceContext
	brainrot.UnimplementedArticleServer
}

func NewArticleServer(svcCtx *svc.ServiceContext) *ArticleServer {
	return &ArticleServer{
		svcCtx: svcCtx,
	}
}

// Post article
func (s *ArticleServer) PostArticle(ctx context.Context, in *brainrot.PostArticleRequest) (*brainrot.PostArticleResponse, error) {
	l := articlelogic.NewPostArticleLogic(ctx, s.svcCtx)
	return l.PostArticle(in)
}

// Delete article
func (s *ArticleServer) DeleteArticle(ctx context.Context, in *brainrot.DeleteArticleRequest) (*brainrot.DeleteArticleResponse, error) {
	l := articlelogic.NewDeleteArticleLogic(ctx, s.svcCtx)
	return l.DeleteArticle(in)
}

// Add tags
func (s *ArticleServer) AddTags(ctx context.Context, in *brainrot.AddTagsRequest) (*brainrot.AddTagsResponse, error) {
	l := articlelogic.NewAddTagsLogic(ctx, s.svcCtx)
	return l.AddTags(in)
}

// Delete tags
func (s *ArticleServer) DeleteTag(ctx context.Context, in *brainrot.DeleteTagRequest) (*brainrot.DeleteTagResponse, error) {
	l := articlelogic.NewDeleteTagLogic(ctx, s.svcCtx)
	return l.DeleteTag(in)
}

// Refresh all articles
func (s *ArticleServer) RefreshAllArticles(ctx context.Context, in *brainrot.RefreshAllArticlesRequest) (*brainrot.RefreshAllArticlesResponse, error) {
	l := articlelogic.NewRefreshAllArticlesLogic(ctx, s.svcCtx)
	return l.RefreshAllArticles(in)
}
