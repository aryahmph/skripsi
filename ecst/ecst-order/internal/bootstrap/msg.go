package bootstrap

import (
	"ecst-order/internal/consts"
	"ecst-order/pkg/logger"
	"ecst-order/pkg/msg"
)

func RegistryMessage() {
	err := msg.Setup("msg.yaml", consts.ConfigPath)
	if err != nil {
		logger.Fatal(logger.SetMessageFormat("file message multi language load error %s", err.Error()))
	}

}
