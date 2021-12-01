package play

import (
	"encoding/json"
	"github.com/bytelang/kplayer/module"
	"github.com/bytelang/kplayer/module/output/provider"
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

func (m AppModule) InitConfig(ctx *kptypes.ClientContext, data json.RawMessage) (interface{}, error) {
	var cfg config.Output
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	m.InitModule(ctx, &cfg)

	return cfg, nil
}

func (m AppModule) ValidateConfig() error {
	return m.Provider.ValidateConfig()
}

func (m AppModule) TriggerMessage(message *kpproto.KPMessage) {
	m.Trigger(message)
}

func (m AppModule) BeginRunning() {
	go m.Provider.StartReconnect()
}

func (m AppModule) EndRunning() {
}
