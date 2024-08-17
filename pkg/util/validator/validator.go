package validator

import (
	"regexp"
)

var reEmail = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,24}$`)

// IsEmail 验证字符串是否是有效的邮箱格式
func IsEmail(email string) bool {
	return reEmail.MatchString(email)
}
