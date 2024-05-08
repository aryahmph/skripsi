package ucase

import (
	"ecst-ticket/internal/appctx"
	"ecst-ticket/internal/consts"
	"ecst-ticket/internal/ucase/contract"
)

type healthCheck struct {
}

func (u *healthCheck) CheckFeatureFlag(data *appctx.Data) bool {
	return true
}

func NewHealthCheck() contract.UseCase {
	return &healthCheck{}
}

func (u *healthCheck) Serve(*appctx.Data) appctx.Response {
	return appctx.Response{
		Name:    consts.ResponseSuccess,
		Message: "ok",
	}
}
