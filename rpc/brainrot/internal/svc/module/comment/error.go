package comment

import (
	"brainrot/pkg/merror"
	"brainrot/rpc/brainrot/internal/svc/module"
)

var (
	ErrCopierCopy = merror.DefineError(merror.Common, module.CommentModuleNumber, 1, "服务器正忙", "copier.Copy 拷贝数据失败")
	ErrDBError    = merror.DefineError(merror.Common, module.CommentModuleNumber, 2, "服务器正忙", "数据库错误")
	ErrAIError    = merror.DefineError(merror.Common, module.CommentModuleNumber, 3, "服务器正忙", "字符和数字间转换错误")

	ErrInvalidInput         = merror.DefineError(merror.Client, module.CommentModuleNumber, 1, "无效输入")
	ErrLackNecessaryField   = merror.DefineError(merror.Client, module.CommentModuleNumber, 2, "缺少必要输入")
	ErrNoPermission         = merror.DefineError(merror.Client, module.CommentModuleNumber, 3, "无权限")
	ErrNeedEnoughReputation = merror.DefineError(merror.Client, module.CommentModuleNumber, 4, "声望不足，至少需要 5 声望")

	ErrSystemError = merror.DefineError(merror.System, module.CommentModuleNumber, 1, "服务器正忙", "操作系统或服务外软件错误")
)
