package server

import (
	"github.com/bytelang/kplayer/module"
)

const AUTHORIZATION_METADATA_KEY = "Authorization"

type ServerCreator interface {
	StartServer(stopChan chan bool, mm module.ModuleManager, authOn bool, authToken string)
}
