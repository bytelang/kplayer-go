package app

import (
    "github.com/bytelang/kplayer/module"
    outputm "github.com/bytelang/kplayer/module/output"
    playm "github.com/bytelang/kplayer/module/play"
    resourcem "github.com/bytelang/kplayer/module/resource"
    "github.com/spf13/cobra"
)

const appName = "kplayer"

var (
    DefaultHome   string
    ModuleManager = newModuleManager()
)

func newModuleManager() module.ModuleManager {
    return module.NewModuleManager(
        playm.NewAppModule(),
        outputm.NewAppModule(),
        resourcem.NewAppModule(),
    )
}

func AddCommands(rootCmd *cobra.Command) {
    for _, m := range ModuleManager {
        if cmd := m.GetCommand(); cmd != nil {
            rootCmd.AddCommand(cmd)
        }
    }
}
