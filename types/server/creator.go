package server

import (
	"github.com/bytelang/kplayer/module"
)

type ServerCreator interface {
	StartServer(stopChan chan bool, mm module.ModuleManager, authOn bool, authToken string)
}
