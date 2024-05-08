package bootstrap

import (
	"ecst-ticket/internal/consts"
	"ecst-ticket/pkg/logger"
	"ecst-ticket/pkg/msg"
)

func RegistryMessage() {
	err := msg.Setup("msg.yaml", consts.ConfigPath)
	if err != nil {
		logger.Fatal(logger.SetMessageFormat("file message multi language load error %s", err.Error()))
	}

}
