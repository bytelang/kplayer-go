package provider

import (
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/output/types"
    kptypes "github.com/bytelang/kplayer/types"
)

type Provider struct {
    config types.Config
    module.ModuleKeeper
}

func NewProvider() *Provider {
    return &Provider{}
}

func (p *Provider) SetConfig(config types.Config) {
    p.config = config
}

func (p *Provider) InitModuleConfig(ctx kptypes.ClientContext, config types.Config) {
    p.SetConfig(config)
}
