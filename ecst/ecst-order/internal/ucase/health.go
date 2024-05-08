package ucase

import (
	"ecst-order/internal/appctx"
	"ecst-order/internal/consts"
	"ecst-order/internal/ucase/contract"
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
