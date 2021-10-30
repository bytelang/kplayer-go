package provider

import (
    "github.com/bytelang/kplayer/module"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    svrproto "github.com/bytelang/kplayer/types/server"
)

type ProviderI interface {
    PluginAdd(plugin *svrproto.PluginAddArgs) (*svrproto.PluginAddReplay, error)
    PluginRemove(plugin *svrproto.PluginRemoveArgs) (*svrproto.PluginRemoveReply, error)
    PluginList(plugin *svrproto.PluginListArgs) (*svrproto.PluginListReply, error)
    PluginUpdate(plugin *svrproto.PluginUpdateArgs) (*svrproto.PluginUpdateReply, error)
}

type Provider struct {
    module.ModuleKeeper
    config config.Plugin
}

var _ ProviderI = &Provider{}

func NewProvider() *Provider {
    return &Provider{}
}

func (p *Provider) SetConfig(config config.Plugin) {
    p.config = config
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config config.Plugin) {
    p.SetConfig(config)
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
}

func (p *Provider) ValidateConfig() error {
    return nil
}
