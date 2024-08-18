package module

import "github.com/m4n5ter/brainrot/pkg/merror"

var (
	PingModuleNumber = merror.MustRegisterErrorModule(0, "Ping")
	UserModuleNumber = merror.MustRegisterErrorModule(1, "User")
)
