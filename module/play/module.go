package play

import (
    "encoding/json"
    "github.com/bytelang/kplayer/module/play/provider"
    "github.com/bytelang/kplayer/module/play/types"
    kpproto "github.com/bytelang/kplayer/proto"
    types2 "github.com/bytelang/kplayer/types"
    "github.com/spf13/cobra"
)

type AppModule struct {
    provider provider.Provider
}

func NewAppModule(provider provider.Provider) AppModule {
    return AppModule{provider: provider}
}

func (m AppModule) GetModuleName() string {
    return types.ModuleName
}

func (m AppModule) GetCommand() *cobra.Command {
    return provider.GetCommand()
}

func (m AppModule) InitConfig(ctx types2.ClientContext, data json.RawMessage) {
    var config types.Config
    if err := json.Unmarshal(data, &config); err != nil {
        panic(err)
    }

    provider.InitConfig(ctx, m.provider, config)
}

func (m AppModule) ParseMessage(message *kpproto.KPMessage) error {
    return m.provider.ParseMessage(message)
}
