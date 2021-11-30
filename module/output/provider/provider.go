package provider

import (
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	kpmsg "github.com/bytelang/kplayer/types/core/proto/msg"
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
	defer p.Trigger(message)

	switch message.Action {
	case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
		p.addOutput()
	case kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD:
		msg := &kpmsg.EventMessageOutputAdd{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		if msg.Error != nil {
			log.WithFields(log.Fields{
				"unique": string(msg.Output.Unique),
				"path":   string(msg.Output.Path),
				"error":  string(msg.Error)}).
				Error("output add failed.")
			return
		}
		log.WithFields(log.Fields{
			"unique": string(msg.Output.Unique),
			"path":   string(msg.Output.Path),
			"error":  string(msg.Error)}).
			Info("output add success.")
	}
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
