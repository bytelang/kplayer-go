package module

import (
    "encoding/json"
    kpproto "github.com/bytelang/kplayer/proto"
    "github.com/bytelang/kplayer/types"
    "github.com/spf13/cobra"
)

type AppModule interface {
    GetModuleName() string
    GetCommand() *cobra.Command
    InitConfig(ctx types.ClientContext, data json.RawMessage)
    ParseMessage(message *kpproto.KPMessage) error
}

type ModuleManager map[string]AppModule

func NewModuleManager(modules ...AppModule) ModuleManager {
    moduleMap := make(ModuleManager)
    for _, module := range modules {
        moduleMap[module.GetModuleName()] = module
    }

    return moduleMap
}
