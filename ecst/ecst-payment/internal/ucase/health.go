package ucase

import (
	"ecst-payment/internal/appctx"
	"ecst-payment/internal/consts"
	"ecst-payment/internal/ucase/contract"
)

type healthCheck struct {
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
