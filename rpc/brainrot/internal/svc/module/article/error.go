package article

import (
	"brainrot/pkg/merror"
	"brainrot/rpc/brainrot/internal/svc/module"
)

var (
	ErrCopierCopy = merror.DefineError(merror.Common, module.ArticleModuleNumber, 1, "服务器正忙", "copier.Copy 拷贝数据失败")
	ErrDBError    = merror.DefineError(merror.Common, module.ArticleModuleNumber, 2, "服务器正忙", "数据库错误")
	ErrAIError    = merror.DefineError(merror.Common, module.ArticleModuleNumber, 3, "服务器正忙", "字符和数字间转换错误")

	ErrInvalidInput       = merror.DefineError(merror.Client, module.ArticleModuleNumber, 1, "无效输入")
	ErrLackNecessaryField = merror.DefineError(merror.Client, module.ArticleModuleNumber, 2, "缺少必要输入")

	ErrSystemError = merror.DefineError(merror.System, module.ArticleModuleNumber, 1, "服务器正忙", "操作系统或服务外软件错误")
)
