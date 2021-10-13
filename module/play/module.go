package play

import (
    "encoding/json"
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/play/provider"
    "github.com/bytelang/kplayer/module/play/types"
    kptypes "github.com/bytelang/kplayer/types"
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

    m.InitModuleConfig(ctx, config)
}
