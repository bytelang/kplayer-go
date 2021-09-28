package player

import (
    "encoding/json"
    "github.com/bytelang/kplayer/client"
    "github.com/bytelang/kplayer/cmd/module/player/provider"
    "github.com/bytelang/kplayer/cmd/module/player/types"
    kpproto "github.com/bytelang/kplayer/proto"
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

func (m AppModule) InitConfig(ctx client.ClientContext, data json.RawMessage) {
    var config types.Config
    if err := json.Unmarshal(data, &config); err != nil {
        panic(err)
    }

    provider.InitConfig(ctx, m.provider, config)
}

func (m AppModule) ParseMessage(ctx client.ClientContext, message *kpproto.KPMessage) {
    m.provider.ParseMessage(ctx, message)
}
