package user

import (
	"brainrot/pkg/merror"
	"brainrot/rpc/brainrot/internal/svc/module"
)

var (
	ErrCopierCopy  = merror.DefineError(merror.Common, module.UserModuleNumber, 1, "服务器正忙", "copier.Copy 拷贝数据失败")
	ErrDBError     = merror.DefineError(merror.Common, module.UserModuleNumber, 2, "服务器正忙", "数据库错误")
	ErrAIError     = merror.DefineError(merror.Common, module.UserModuleNumber, 3, "服务器正忙", "字符和数字间转换错误")
	ErrServerError = merror.DefineError(merror.Common, module.UserModuleNumber, 4, "服务器正忙", "服务端错误")

	ErrInvalidInput               = merror.DefineError(merror.Client, module.UserModuleNumber, 1, "无效输入")
	ErrLackNecessaryField         = merror.DefineError(merror.Client, module.UserModuleNumber, 2, "缺少必要输入")
	ErrInvalidRefreshToken        = merror.DefineError(merror.Client, module.UserModuleNumber, 3, "无效的令牌", "无效的刷新令牌")
	ErrExpiredRefreshToken        = merror.DefineError(merror.Client, module.UserModuleNumber, 4, "无效的令牌", "刷新令牌已过期")
	ErrDBUserNotFound             = merror.DefineError(merror.Client, module.UserModuleNumber, 5, "用户不存在", "数据库中未找到用户")
	ErrDBDuplicateUsernameOrEmail = merror.DefineError(merror.Client, module.UserModuleNumber, 6, "用户名或邮箱已存在", "数据库中已存在相同用户名或邮箱")

	ErrSystemError = merror.DefineError(merror.System, module.UserModuleNumber, 1, "服务器正忙", "操作系统或服务外软件错误")
)
