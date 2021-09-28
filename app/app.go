package app

import (
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/play"
    "github.com/bytelang/kplayer/module/play/provider"
    "github.com/spf13/cobra"
)

const appName = "kplayer"

var (
    DefaultHome   string
    ModuleManager = newKplayerApp()
)

type KplayerApp struct {
    Manager module.ModuleManager
}

func newKplayerApp() *KplayerApp {
    app := &KplayerApp{}

    playerProvider := provider.NewProvider()
    app.Manager = module.NewModuleManager(
        play.NewAppModule(playerProvider),
    )

    return app
}

func (ka KplayerApp) AddCommands(rootCmd *cobra.Command) {
    for _, m := range ka.Manager {
        if cmd := m.GetCommand(); cmd != nil {
            rootCmd.AddCommand(cmd)
        }
    }
}
