package app

import (
    "github.com/bytelang/kplayer/client"
    "github.com/bytelang/kplayer/cmd/module/player"
    "github.com/bytelang/kplayer/cmd/module/player/provider"
)

const appName = "kplayer"

var (
    DefaultHome string
    ModuleBasic = client.NewModuleManager(
        player.AppModule{},
    )
)

type KplayerApp struct {
    mm *client.ModuleManager
}

func NewKplayerApp() *KplayerApp {
    app := &KplayerApp{}

    playerProvider := provider.NewProvider()
    app.mm = client.NewModuleManager(
        player.NewAppModule(playerProvider),
    )

    return app
}
