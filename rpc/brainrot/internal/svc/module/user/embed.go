package user

import (
	_ "embed"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	//go:embed hsetnxscript.lua
	setLuaScript string
	SetScript    = redis.NewScript(setLuaScript)
)
