package user

import (
	"github.com/m4n5ter/brainrot/pkg/merror"
	"github.com/m4n5ter/brainrot/rpc/brainrot/internal/svc/module"
)

var (
	ErrCopierCopy = merror.DefineError(merror.Common, module.UserModuleNumber, 1, "服务器正忙", "copier.Copy 拷贝数据失败")
	ErrDBError    = merror.DefineError(merror.Common, module.UserModuleNumber, 2, "服务器正忙", "数据库错误")

	ErrInvalidInput        = merror.DefineError(merror.Client, module.UserModuleNumber, 1, "无效输入")
	ErrLackNecessaryField  = merror.DefineError(merror.Client, module.UserModuleNumber, 2, "缺少必要输入")
	ErrInvalidRefreshToken = merror.DefineError(merror.Client, module.UserModuleNumber, 3, "无效的令牌", "无效的刷新令牌")
	ErrExpiredRefreshToken = merror.DefineError(merror.Client, module.UserModuleNumber, 4, "无效的令牌", "刷新令牌已过期")

	ErrSystemError = merror.DefineError(merror.System, module.UserModuleNumber, 1, "服务器正忙", "操作系统错误")
)
