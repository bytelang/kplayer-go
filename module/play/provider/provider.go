package provider

import (
    "github.com/bytelang/kplayer/module"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
)

// Provider play module provider
type Provider struct {
    config *config.Play
    module.ModuleKeeper
}

type ProviderI interface {
    GetStartPoint() uint32
    GetPlayModel() string
}

// NewProvider return provider
func NewProvider() *Provider {
    return &Provider{
        config: &config.Play{},
    }
}

func (p *Provider) GetConfig() *config.Play {
    return p.config
}

func (p *Provider) setConfig(config config.Play) {
    p.config = &config
}

// InitConfig set module config on kplayer started
func (p *Provider) InitModuleConfig(ctx *kptypes.ClientContext, config config.Play) {
    p.setConfig(config)
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
}

func (p *Provider) ValidateConfig() error {
    return nil
}
