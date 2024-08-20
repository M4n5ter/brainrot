package module

import "brainrot/pkg/merror"

var (
	PingModuleNumber    = merror.MustRegisterErrorModule(0, "Ping")
	UserModuleNumber    = merror.MustRegisterErrorModule(1, "User")
	ArticleModuleNumber = merror.MustRegisterErrorModule(2, "Article")
)
