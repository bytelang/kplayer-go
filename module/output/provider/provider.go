package provider

import (
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	kpmsg "github.com/bytelang/kplayer/types/core/proto/msg"
	"github.com/bytelang/kplayer/types/core/proto/prompt"
	moduletypes "github.com/bytelang/kplayer/types/module"
	svrproto "github.com/bytelang/kplayer/types/server"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ProviderI interface {
	OutputAdd(output *svrproto.OutputAddArgs) (*svrproto.OutputAddReply, error)
	OutputRemove(output *svrproto.OutputRemoveArgs) (*svrproto.OutputRemoveReply, error)
	OutputList(output *svrproto.OutputListArgs) (*svrproto.OutputListReply, error)
}

type Provider struct {
	module.ModuleKeeper

	// module outputs
	outputs           Outputs
	reconnectInternal int32

	// reconnect
	reconnectChan chan interface{}
	reconnectWait sync.WaitGroup
}

var _ ProviderI = &Provider{}

func NewProvider() *Provider {
	return &Provider{
		reconnectChan: make(chan interface{}, 5),
	}
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config *config.Output, homePath string) {
	// set module attribute
	p.reconnectInternal = config.ReconnectInternal

	for _, item := range config.Lists {
		unique := item.Unique
		if unique == "" {
			unique = kptypes.GetRandString(6)
		}

		p.outputs = append(p.outputs, moduletypes.Output{
			Path:       item.Path,
			Unique:     unique,
			CreateTime: uint64(time.Now().Unix()),
			StartTime:  0,
			EndTime:    0,
			Connected:  false,
		})
	}
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
	defer p.Trigger(message)

	switch message.Action {
	case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
		p.addOutputList()
	case kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD:
		msg := &kpmsg.EventMessageOutputAdd{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)

		logFields := log.WithFields(log.Fields{
			"unique": string(msg.Output.Unique),
			"path":   string(msg.Output.Path)})
		if string(msg.Error) != "" {
			logFields = logFields.WithField("error", string(msg.Error))
		}

		if msg.Error != nil {
			logFields.Error("output add failed")

			// send reconnect instance to channel
			if p.reconnectInternal > 0 {
				p.reconnectChan <- config.OutputInstance{
					Path:   string(msg.Output.Path),
					Unique: string(msg.Output.Unique),
				}
			}
			return
		}

		logFields.Info("output add success")

		// update output status
		if output := p.outputs.GetOutputByUnique(string(msg.Output.Unique)); output != nil {
			output.StartTime = uint64(time.Now().Unix())
		}
	case kpproto.EVENT_MESSAGE_ACTION_OUTPUT_DISCONNECT:
		msg := &kpmsg.EventMessageOutputDisconnect{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)

		logFields := log.WithFields(log.Fields{
			"unique": string(msg.Output.Unique),
			"path":   string(msg.Output.Path),
			"error":  string(msg.Error)})
		logFields.Error("output disconnection")

		// send reconnect instance to channel
		if p.reconnectInternal > 0 {
			p.reconnectChan <- config.OutputInstance{
				Path:   string(msg.Output.Path),
				Unique: string(msg.Output.Unique),
			}
		}

		// update output status
		if output := p.outputs.GetOutputByUnique(string(msg.Output.Unique)); output != nil {
			output.EndTime = uint64(time.Now().Unix())
		}
	case kpproto.EVENT_MESSAGE_ACTION_PLAYER_ENDED:
		p.reconnectChan <- nil
	}
}

func (p *Provider) ValidateConfig() error {
	return nil
}

func (p *Provider) addOutputList() {
	for _, item := range p.outputs {
		if err := p.addOutput(item.Path, item.Unique); err != nil {
			log.Warn(err)
		}
	}
}

func (p *Provider) addOutput(path string, unique string) error {
	corePlayer := core.GetLibKplayerInstance()

	if err := corePlayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_OUTPUT_ADD, &prompt.EventPromptOutputAdd{
		Output: &kpproto.PromptOutput{
			Path:   []byte(path),
			Unique: []byte(unique),
		},
	}); err != nil {
		log.Warn(err)
	}

	return nil
}

func (p *Provider) StartReconnect() {
	p.reconnectWait.Add(1)
	defer p.reconnectWait.Done()

	for {
		instance := <-p.reconnectChan

		switch instance.(type) {
		case config.OutputInstance:
			ins := instance.(config.OutputInstance)
			logFields := log.WithFields(log.Fields{"path": ins.Path, "unique": ins.Unique})

			logFields.Infof("will be reconnect on after %d seconds", p.reconnectInternal)
			time.Sleep(time.Second * time.Duration(p.reconnectInternal))

			if err := p.addOutput(ins.Path, ins.Unique); err != nil {
				logFields.Warn("reconnect output failed. it will not try again")
			}
		case nil:
			log.Debug("reconnect coroutine stop")
			return
		default:
			log.Fatalf("invalid type")
		}
	}
}
