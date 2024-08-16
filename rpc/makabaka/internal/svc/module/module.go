package module

import "github.com/m4n5ter/makabaka/pkg/merror"

var (
	PingModuleNumber = merror.MustRegisterErrorModule(0, "Ping")
	UserModuleNumber = merror.MustRegisterErrorModule(1, "User")
)
