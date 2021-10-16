package provider

import (
    "github.com/bytelang/kplayer/core"
    "github.com/bytelang/kplayer/module"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    kpproto "github.com/bytelang/kplayer/types/core"
    "github.com/bytelang/kplayer/types/core/prompt"
    log "github.com/sirupsen/logrus"
)

type Provider struct {
    module.ModuleKeeper
    config config.Output
}

func NewProvider() *Provider {
    return &Provider{}
}

func (p *Provider) SetConfig(config config.Output) {
    p.config = config
}

func (p *Provider) InitModuleConfig(ctx *kptypes.ClientContext, config config.Output) {
    p.SetConfig(config)
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
    switch message.Action {
    case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
        p.addOutput()
    }

    p.Trigger(message)
}

func (p *Provider) ValidateConfig() error {
    return nil
}

func (p *Provider) addOutput() {
    corePlayer := core.GetLibKplayerInstance()
    for _, item := range p.config.Lists {
        if err := corePlayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_OUTPUT_ADD, &prompt.EventPromptOutputAdd{
            Path:   []byte(item.Path),
            Unique: []byte(item.Unique),
        }); err != nil {
            log.Warn(err)
        }
    }
}
