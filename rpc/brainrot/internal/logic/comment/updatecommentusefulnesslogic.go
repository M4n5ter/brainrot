package commentlogic

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	commentmodule "brainrot/rpc/brainrot/internal/svc/module/comment"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"google.golang.org/grpc/metadata"
)

type UpdateCommentUsefulnessLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateCommentUsefulnessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateCommentUsefulnessLogic {
	return &UpdateCommentUsefulnessLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Update comment usefulness
func (l *UpdateCommentUsefulnessLogic) UpdateCommentUsefulness(in *brainrot.UpdateCommentUsefulnessRequest) (*brainrot.UpdateCommentUsefulnessResponse, error) {
	if in.CommentId == 0 {
		return nil, commentmodule.ErrLackNecessaryField
	}

	modelcomment, err := l.svcCtx.CommentModel.FindOne(l.ctx, in.CommentId)
	if err != nil {
		return nil, commentmodule.ErrInvalidInput.Wrap("评论 %d 不存在", in.CommentId)
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

	voterIDs := strings.Split(modelcomment.VoterIds, ",")
	for _, voterID := range voterIDs {
		if useridstr == strings.TrimSpace(voterID) {
			return nil, commentmodule.ErrVoteTwice
		}
	}
	voterIDs = append(voterIDs, useridstr)
	voterIDsStr := strings.Join(voterIDs, ",")

	modeluser, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userid))
	if err != nil {
		return nil, commentmodule.ErrDBError.Wrap("查找用户 %d 失败", userid)
	}

	if modeluser.Reputation < 5 {
		return nil, commentmodule.ErrNeedEnoughReputation
	}

	err = l.svcCtx.CommentModel.Trans(l.ctx, func(ctx context.Context, session sqlx.Session) error {
		commentQuery := fmt.Sprintf("UPDATE `comment` SET `useful_count` = `useful_count` + 1, `voter_ids` = `%s` WHERE `id` = ? AND `status` = 1", voterIDsStr)
		if !in.IsUseful {
			commentQuery = fmt.Sprintf("UPDATE `comment` SET `useless_count` = `useless_count` + 1, `voter_ids` = `%s` WHERE `id` = ? AND `status` = 1", voterIDsStr)
		}
		_, err := session.ExecCtx(ctx, commentQuery, in.CommentId)
		if err != nil {
			return err
		}

		_, err = session.ExecCtx(ctx, "UPDATE `user` SET `reputation` = `reputation` - 5 WHERE `id` = ? AND `status` = 1", modeluser.Id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, commentmodule.ErrDBError.Wrap("事务失败，更新评论 %d 和用户 %d 失败", in.CommentId, userid)
	}

	if in.IsUseful {
		modelcomment.UsefulCount++
	} else {
		modelcomment.UselessCount++
	}

	return &brainrot.UpdateCommentUsefulnessResponse{UsefulCount: modelcomment.UsefulCount, UselessCount: modelcomment.UselessCount}, nil
}
