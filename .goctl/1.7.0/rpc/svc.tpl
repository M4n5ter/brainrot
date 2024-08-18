package svc

import {{.imports}}
import merror "github.com/m4n5ter/brainrot/pkg/error"

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:c,
	}
}