package provider

import (
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/resource/types"
    kpproto "github.com/bytelang/kplayer/proto"
    kptypes "github.com/bytelang/kplayer/types"
    log "github.com/sirupsen/logrus"
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

func (p Provider) ParseMessage(message *kpproto.KPMessage) error {
    switch message.Action {
    case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
        log.Info("kplayer success run")
    }

    return nil
}

func (p *Provider) InitModuleConfig(ctx kptypes.ClientContext, config types.Config) {
    p.SetConfig(config)
}
