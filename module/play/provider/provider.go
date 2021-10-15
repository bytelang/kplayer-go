package provider

import (
    "github.com/bytelang/kplayer/module"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
)

// Provider play module provider
type Provider struct {
    config config.Play
    module.ModuleKeeper
}

// NewProvider return provider
func NewProvider() *Provider {
    return &Provider{}
}

func (p *Provider) setConfig(config config.Play) {
    p.config = config
}

// InitConfig set module config on kplayer started
func (p *Provider) InitModuleConfig(ctx kptypes.ClientContext, config config.Play) {
    p.setConfig(config)
}
