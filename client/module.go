package client

import (
    "encoding/json"
    kpproto "github.com/bytelang/kplayer/proto"
    "github.com/spf13/cobra"
)

type AppModule interface {
    GetModuleName() string
    GetCommand() *cobra.Command
    InitConfig(ctx ClientContext, data json.RawMessage)
    ParseMessage(ctx ClientContext, message *kpproto.KPMessage)
}

type ModuleManager map[string]AppModule

func NewModuleManager(modules ...AppModule) *ModuleManager {
    moduleMap := make(ModuleManager)
    for _, module := range modules {
        moduleMap[module.GetModuleName()] = module
    }

    return &moduleMap
}

func (bm ModuleManager) AddCommands(rootCmd *cobra.Command) {
    for _, m := range bm {
        if cmd := m.GetCommand(); cmd != nil {
            rootCmd.AddCommand(cmd)
        }
    }
}
