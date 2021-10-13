package provider

import (
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/play/types"
    kpproto "github.com/bytelang/kplayer/proto"
    "github.com/bytelang/kplayer/proto/msg"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/golang/protobuf/proto"
    log "github.com/sirupsen/logrus"
)

// Provider play module provider
type Provider struct {
    config types.Config
    module.ModuleKeeper
}

// NewProvider return provider
func NewProvider() Provider {
    return Provider{}
}

func (p *Provider) setConfig(config types.Config) {
    p.config = config
}

// ParseMessage handle core message event
func (p Provider) ParseMessage(message *kpproto.KPMessage) error {
    switch message.Action {
    case kpproto.EventAction_EVENT_MESSAGE_ACTION_PLAYER_STARTED:
        log.Info("Core success run")
    case kpproto.EventAction_EVENT_MESSAGE_ACTION_RESOURCE_REMOVE:
        resourceMsg := &msg.EventMessageResourceRemove{}
        if err := proto.Unmarshal([]byte(message.Body), resourceMsg); err != nil {
            return err
        }

        p.Trigger(message.Action, resourceMsg)
    }

    return nil
}

// InitConfig set module config on kplayer started
func (p *Provider) InitConfig(ctx kptypes.ClientContext, config types.Config) {
    p.setConfig(config)
}
