package play

import (
    "encoding/json"
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/plugin/provider"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    "github.com/spf13/cobra"
)

type AppModule struct {
    *provider.Provider
}

var _ module.AppModule = &AppModule{}

func NewAppModule() AppModule {
    return AppModule{provider.NewProvider()}
}

func (m AppModule) GetModuleName() string {
    return provider.ModuleName
}

func (m AppModule) GetCommand() *cobra.Command {
    return provider.GetCommand()
}

func (m AppModule) InitConfig(ctx *kptypes.ClientContext, data json.RawMessage) error {
    var cfg config.Plugin
    if err := json.Unmarshal(data, &cfg); err != nil {
        return err
    }

    m.InitModule(ctx, cfg)

    return nil
}

func (m AppModule) ValidateConfig() error {
    return m.Provider.ValidateConfig()
}

func (m AppModule) TriggerMessage(message *kpproto.KPMessage) {
    m.Trigger(message)
}