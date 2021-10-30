package provider

import (
    "github.com/bytelang/kplayer/core"
    "github.com/bytelang/kplayer/module"
    kptypes "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    "github.com/bytelang/kplayer/types/core/proto/prompt"
    svrproto "github.com/bytelang/kplayer/types/server"
    log "github.com/sirupsen/logrus"
)

type ProviderI interface {
    OutputAdd(output *svrproto.OutputAddArgs) (*svrproto.OutputAddReply, error)
    OutputRemove(output *svrproto.OutputRemoveArgs) (*svrproto.OutputRemoveReply, error)
    OutputList(output *svrproto.OutputListArgs) (*svrproto.OutputListReply, error)
}

type Provider struct {
    module.ModuleKeeper
    config config.Output
}

var _ ProviderI = &Provider{}

func NewProvider() *Provider {
    return &Provider{}
}

func (p *Provider) SetConfig(config config.Output) {
    p.config = config
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config config.Output) {
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
            Output: &kpproto.PromptOutput{
                Path:   []byte(item.Path),
                Unique: []byte(item.Unique),
            },
        }); err != nil {
            log.Warn(err)
        }
    }
}
