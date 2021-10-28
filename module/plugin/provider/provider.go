package provider

import (
    "github.com/bytelang/kplayer/module"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
)

type Provider struct {
    module.ModuleKeeper
    config config.Plugin
}

func NewProvider() *Provider {
    return &Provider{}
}

func (p *Provider) SetConfig(config config.Plugin) {
    p.config = config
}

func (p *Provider) InitModuleConfig(ctx *kptypes.ClientContext, config config.Plugin) {
    p.SetConfig(config)
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
}

func (p *Provider) ValidateConfig() error {
    return nil
}
