package provider

import (
    "github.com/bytelang/kplayer/module"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
)

type Provider struct {
    config config.Output
    module.ModuleKeeper
}

func NewProvider() *Provider {
    return &Provider{}
}

func (p *Provider) SetConfig(config config.Output) {
    p.config = config
}

func (p *Provider) InitModuleConfig(ctx kptypes.ClientContext, config config.Output) {
    p.SetConfig(config)
}
