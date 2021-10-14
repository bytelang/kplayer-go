package provider

import (
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/play/types"
    kptypes "github.com/bytelang/kplayer/types"
)

// Provider play module provider
type Provider struct {
    config types.Config
    module.ModuleKeeper
}

// NewProvider return provider
func NewProvider() *Provider {
    return &Provider{}
}

func (p *Provider) setConfig(config types.Config) {
    p.config = config
}

// InitConfig set module config on kplayer started
func (p *Provider) InitModuleConfig(ctx kptypes.ClientContext, config types.Config) {
    p.setConfig(config)
}
