package commentlogic

import (
	"context"
	"strconv"

	"brainrot/gen/pb/brainrot"
	"brainrot/model"
	"brainrot/rpc/brainrot/internal/svc"
	commentmodule "brainrot/rpc/brainrot/internal/svc/module/comment"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type PostCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewPostCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PostCommentLogic {
	return &PostCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Post comment
func (l *PostCommentLogic) PostComment(in *brainrot.PostCommentRequest) (*brainrot.PostCommentResponse, error) {
	if in.ArticleId == 0 || in.Content == "" {
		return nil, commentmodule.ErrLackNecessaryField
	}

	ids := metadata.ValueFromIncomingContext(l.ctx, "userid")
	if ids == nil {
		return nil, commentmodule.ErrSystemError.Wrap("元数据中不存在 userid")
	}

	useridstr := ids[0]
	userid, err := strconv.Atoi(useridstr)
	if err != nil {
		return nil, commentmodule.ErrAIError
	}

	modeluser, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userid))
	if err != nil {
		return nil, commentmodule.ErrDBError.Wrap("查找用户 %d 失败", userid)
	}

	result, err := l.svcCtx.CommentModel.Insert(l.ctx, &model.Comment{
		ArticleId: in.ArticleId,
		UserId:    uint64(userid),
		Commenter: modeluser.Username,
		Content:   in.Content,
	})
	if err != nil {
		return nil, commentmodule.ErrDBError.Wrap("插入评论失败：%v", err)
	}

	commentID, err := result.LastInsertId()
	if err != nil {
		return nil, commentmodule.ErrDBError.Wrap("获取评论ID失败：%v", err)
	}

	return &brainrot.PostCommentResponse{CommentId: uint64(commentID)}, nil
}
