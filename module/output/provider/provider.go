package provider

import (
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/output/types"
    kpproto "github.com/bytelang/kplayer/proto"
    "github.com/bytelang/kplayer/proto/msg"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/gogo/protobuf/proto"
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

func (p *Provider) ParseMessage(message *kpproto.KPMessage) error {
    switch message.Action {
    case kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD:
        outputMsg := &msg.EventMessageOutputAdd{}
        if err := proto.Unmarshal([]byte(message.Body), outputMsg); err != nil {
            return err
        }
        p.Trigger(message.Action, outputMsg)
    }

    return nil
}

func (p *Provider) InitModuleConfig(ctx kptypes.ClientContext, config types.Config) {
    p.SetConfig(config)
}
