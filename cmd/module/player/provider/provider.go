package provider

import (
    "github.com/bytelang/kplayer/client"
    "github.com/bytelang/kplayer/cmd/module/player/types"
    kpproto "github.com/bytelang/kplayer/proto"
    log "github.com/sirupsen/logrus"
)

type Provider struct {
    config types.Config
}

func NewProvider() Provider {
    return Provider{}
}

func (p *Provider) SetConfig(config types.Config) {
    p.config = config
}

func (p Provider) ParseMessage(ctx client.ClientContext, message *kpproto.KPMessage) {
    switch message.Action {
    case kpproto.EventAction_EVENT_MESSAGE_ACTION_PLAYER_STARTED:
        log.Info("kplayer success run.")
    }
}
