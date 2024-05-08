package bootstrap

import (
	"en-ticket/internal/consts"
	"en-ticket/pkg/logger"
	"en-ticket/pkg/msg"
)

func RegistryMessage() {
	err := msg.Setup("msg.yaml", consts.ConfigPath)
	if err != nil {
		logger.Fatal(logger.SetMessageFormat("file message multi language load error %s", err.Error()))
	}

}
