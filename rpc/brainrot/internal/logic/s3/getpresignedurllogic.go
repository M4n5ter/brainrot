package s3logic

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"brainrot/gen/pb/brainrot"
	"brainrot/rpc/brainrot/internal/svc"
	s3module "brainrot/rpc/brainrot/internal/svc/module/s3"

	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/metadata"
)

type GetPresignedURLLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetPresignedURLLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPresignedURLLogic {
	return &GetPresignedURLLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// Get presigned url
func (l *GetPresignedURLLogic) GetPresignedURL(in *brainrot.GetPresignedURLRequest) (*brainrot.GetPresignedURLResponse, error) {
	if in.ObjectKey == "" || in.ContentType == "" {
		return nil, s3module.ErrLackNecessaryField
	}

	ids := metadata.ValueFromIncomingContext(l.ctx, "userid")
	if ids == nil {
		return nil, s3module.ErrSystemError.Wrap("元数据中不存在 userid")
	}

	useridstr := ids[0]
	userid, err := strconv.Atoi(useridstr)
	if err != nil {
		return nil, s3module.ErrAIError
	}

	modeluser, err := l.svcCtx.UserModel.FindOne(l.ctx, uint64(userid))
	if err != nil {
		return nil, s3module.ErrDBError.Wrap("%v", err)
	}

	bucketName, objectName := l.svcCtx.Config.S3.PrivateBucket, fmt.Sprintf("%s/%s", modeluser.Username, in.ObjectKey)
	if in.IsPublic {
		bucketName = l.svcCtx.Config.S3.PublicBucket
	}

	switch strings.ToUpper(in.Operation) {
	case "GET":
		reqParams := make(url.Values)
		presigned, err := l.svcCtx.S3.PresignedGetObject(l.ctx, bucketName, objectName, time.Hour, reqParams)
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("获取 Presigned URL 失败：%v", err)
		}

		return &brainrot.GetPresignedURLResponse{Url: presigned.String()}, nil
	case "PUT":
		presigned, err := l.svcCtx.S3.PresignedPutObject(l.ctx, bucketName, objectName, time.Hour)
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("获取 Presigned URL 失败：%v", err)
		}

		return &brainrot.GetPresignedURLResponse{Url: presigned.String()}, nil
	case "HEAD":
		reqParams := make(url.Values)
		presigned, err := l.svcCtx.S3.PresignedHeadObject(l.ctx, bucketName, objectName, time.Hour, reqParams)
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("获取 Presigned URL 失败：%v", err)
		}

		return &brainrot.GetPresignedURLResponse{Url: presigned.String()}, nil
	case "POST":
		policy := minio.NewPostPolicy()
		err := policy.SetBucket(bucketName)
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("设置 Bucket 失败：%v", err)
		}

		err = policy.SetKey(objectName)
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("设置 ObjectKey 失败：%v", err)
		}

		err = policy.SetExpires(time.Now().UTC().Add(30 * time.Second))
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("设置过期时间失败：%v", err)
		}

		err = policy.SetContentType(in.ContentType)
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("设置 ContentType 失败：%v", err)
		}

		err = policy.SetContentLengthRange(1, 1024*1024*10) // 1B - 10MB
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("设置 ContentLengthRange 失败：%v", err)
		}

		err = policy.SetUserMetadata("userid", useridstr)
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("设置 UserMetadata 失败：%v", err)
		}

		presigned, formData, err := l.svcCtx.S3.PresignedPostPolicy(l.ctx, policy)
		if err != nil {
			return nil, s3module.ErrSystemError.Wrap("获取 Presigned URL 失败：%v", err)
		}

		return &brainrot.GetPresignedURLResponse{Url: presigned.String(), FormData: formData}, nil
	default:
		return nil, s3module.ErrInvalidOperation
	}
}
