package bootstrap

import (
	"en-payment/internal/consts"
	"en-payment/pkg/logger"
	"en-payment/pkg/msg"
)

func RegistryMessage() {
	err := msg.Setup("msg.yaml", consts.ConfigPath)
	if err != nil {
		logger.Fatal(logger.SetMessageFormat("file message multi language load error %s", err.Error()))
	}

}
