package s3

import (
	"brainrot/pkg/merror"
	"brainrot/rpc/brainrot/internal/svc/module"
)

var (
	ErrAIError = merror.DefineError(merror.Common, module.S3ModuleNumber, 1, "服务器正忙", "字符和数字间转换错误")
	ErrDBError = merror.DefineError(merror.Common, module.S3ModuleNumber, 2, "数据库错误", "数据库操作错误")

	ErrInvalidInput       = merror.DefineError(merror.Client, module.S3ModuleNumber, 1, "无效输入")
	ErrLackNecessaryField = merror.DefineError(merror.Client, module.S3ModuleNumber, 2, "缺少必要输入")
	ErrInvalidOperation   = merror.DefineError(merror.Client, module.S3ModuleNumber, 3, "无效操作")

	ErrSystemError = merror.DefineError(merror.System, module.S3ModuleNumber, 1, "服务器正忙", "操作系统或服务外软件错误")
)
