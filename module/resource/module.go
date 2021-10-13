package play

import (
    "encoding/json"
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/resource/provider"
    "github.com/bytelang/kplayer/module/resource/types"
    kpproto "github.com/bytelang/kplayer/proto"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/spf13/cobra"
)

type AppModule struct {
    provider provider.Provider
}

var _ module.AppModule = &AppModule{}

func NewAppModule() AppModule {
    return AppModule{provider: provider.NewProvider()}
}

func (m AppModule) GetModuleName() string {
    return types.ModuleName
}

func (m AppModule) GetCommand() *cobra.Command {
    return provider.GetCommand()
}

func (m AppModule) InitConfig(ctx kptypes.ClientContext, data json.RawMessage) {
    var config types.Config
    if err := json.Unmarshal(data, &config); err != nil {
        panic(err)
    }

    m.provider.InitConfig(ctx, config)
}

func (m AppModule) ParseMessage(message *kpproto.KPMessage) error {
    return m.provider.ParseMessage(message)
}
