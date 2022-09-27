package provider

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	kpmsg "github.com/bytelang/kplayer/types/core/proto/msg"
	kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
	moduletypes "github.com/bytelang/kplayer/types/module"
	svrproto "github.com/bytelang/kplayer/types/server"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type ProviderI interface {
	OutputAdd(ctx context.Context, output *svrproto.OutputAddArgs) (*svrproto.OutputAddReply, error)
	OutputRemove(ctx context.Context, output *svrproto.OutputRemoveArgs) (*svrproto.OutputRemoveReply, error)
	OutputList(ctx context.Context, output *svrproto.OutputListArgs) (*svrproto.OutputListReply, error)
	mustEmbedUnimplementedOutputGreeterServer()
}

func (p Provider) mustEmbedUnimplementedOutputGreeterServer() {
	//TODO implement me
	panic("implement me")
}

type Provider struct {
	module.ModuleKeeper
	svrproto.UnimplementedOutputGreeterServer

	// module outputs
	configList        Outputs
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

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config *config.Output) {
	// set module attribute
	p.reconnectInternal = config.ReconnectInternal

	for _, item := range config.Lists {
		unique := item.Unique
		if unique == "" {
			unique = kptypes.GetUniqueString(item.Path)
		}

		if err := p.configList.AppendOutput(moduletypes.Output{
			Path:       item.Path,
			Unique:     unique,
			CreateTime: uint64(time.Now().Unix()),
			StartTime:  0,
			EndTime:    0,
			Connected:  false,
		}); err != nil {
			log.Fatal(err)
		}
	}
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
	switch message.Action {
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_ADD:
		msg := &kpmsg.EventMessageOutputAdd{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)

		logFields := log.WithFields(log.Fields{
			"unique": msg.Output.Unique,
			"path":   msg.Output.Path})
		if msg.Error != "" {
			logFields = logFields.WithField("error", msg.Error)
		}

		if len(msg.Error) != 0 {
			logFields.Errorf("output add failed. error: %s", msg.Error)

			// send reconnect instance to channel
			if p.reconnectInternal > 0 {
				p.reconnectChan <- config.OutputInstance{
					Path:   msg.Output.Path,
					Unique: msg.Output.Unique,
				}
			}
			return
		}

		logFields.Info("output add success")

		// update output status
		output, _, err := p.configList.GetOutputByUnique(msg.Output.Unique)
		if err != nil {
			logFields.WithField("error", err).Fatal("update output status failed")
		}

		output.StartTime = uint64(time.Now().Unix())
		output.EndTime = 0
		output.Connected = true
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_REMOVE:
		msg := &kpmsg.EventMessageOutputRemove{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		logFields := log.WithFields(log.Fields{
			"unique": msg.Output.Unique,
			"path":   msg.Output.Path,
		})

		if _, err := p.configList.RemoveOutputByUnique(msg.Output.Unique); err != nil {
			logFields.Fatal("remove output failed")
		}

		logFields.Info("remove output success")
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_DISCONNECT:
		msg := &kpmsg.EventMessageOutputDisconnect{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)

		logFields := log.WithFields(log.Fields{
			"unique": msg.Output.Unique,
			"path":   msg.Output.Path,
			"error":  msg.Error})
		logFields.Error("output disconnection")

		// send reconnect instance to channel
		if p.reconnectInternal > 0 {
			p.reconnectChan <- config.OutputInstance{
				Path:   msg.Output.Path,
				Unique: msg.Output.Unique,
			}

			// update output status
			output, _, err := p.configList.GetOutputByUnique(msg.Output.Unique)
			if err != nil {
				logFields.WithField("error", err).Fatal("update output status failed")
			}

			output.EndTime = uint64(time.Now().Unix())
			output.Connected = false
			break
		}

		// remove output
		removeOutput, err := p.configList.RemoveOutputByUnique(msg.Output.Unique)
		logResultFields := log.WithFields(log.Fields{"path": removeOutput.Path, "unique": removeOutput.Unique})
		if err != nil {
			logResultFields.Fatal("remove disconnected output failed")
		}
		logResultFields.Info("remove disconnected output success")
	}
}

func (p *Provider) ValidateConfig() error {
	existName := []string{}
	for _, item := range p.configList.outputs {
		if kptypes.ArrayInString(existName, item.Unique) {
			return OutputUniqueHasExisted
		}

		if item.Path == "" {
			return fmt.Errorf("output path cannot be empty")
		}
		existName = append(existName, item.Unique)
	}

	return nil
}

func (p *Provider) addOutput(output moduletypes.Output) error {
	// validate
	if p.configList.Exist(output.Unique) {
		return OutputUniqueHasExisted
	}

	// send prompt
	corePlayer := core.GetLibKplayerInstance()

	if err := corePlayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_OUTPUT_ADD, &kpprompt.EventPromptOutputAdd{
		Output: &kpprompt.PromptOutput{
			Path:   output.Path,
			Unique: output.Unique,
		},
	}); err != nil {
		log.Warn(err)
	}

	if err := p.configList.AppendOutput(output); err != nil {
		return err
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

			corePlayer := core.GetLibKplayerInstance()
			_ = corePlayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_OUTPUT_ADD, &kpprompt.EventPromptOutputAdd{
				Output: &kpprompt.PromptOutput{
					Path:   ins.Path,
					Unique: ins.Unique,
				},
			})
		case nil:
			log.Debug("reconnect coroutine stop")
			return
		default:
			log.Fatalf("invalid type")
		}
	}
}

func (p *Provider) BeginRunning() {
	for _, item := range p.configList.outputs {
		if err := core.GetLibKplayerInstance().AddOutput(&kpprompt.EventPromptOutputAdd{
			Output: &kpprompt.PromptOutput{
				Path:   item.Path,
				Unique: item.Unique,
			},
		}); err != nil {
			log.WithFields(log.Fields{"unique": item.Unique, "path": item.Path, "error": err}).Trace("add output failed")
		}
	}
}

func (p *Provider) EndReconnect() {
	p.reconnectChan <- nil
}
