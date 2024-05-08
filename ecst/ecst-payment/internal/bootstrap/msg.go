package bootstrap

import (
	"ecst-payment/internal/consts"
	"ecst-payment/pkg/logger"
	"ecst-payment/pkg/msg"
)

func RegistryMessage() {
	err := msg.Setup("msg.yaml", consts.ConfigPath)
	if err != nil {
		logger.Fatal(logger.SetMessageFormat("file message multi language load error %s", err.Error()))
	}

}
