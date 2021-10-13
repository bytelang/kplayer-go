package app

import (
    "github.com/bytelang/kplayer/module"
    playm "github.com/bytelang/kplayer/module/play"
    outputm "github.com/bytelang/kplayer/module/output"
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
