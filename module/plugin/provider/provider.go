package provider

import (
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
	"time"
)

type ProviderI interface {
	PluginAdd(plugin *svrproto.PluginAddArgs) (*svrproto.PluginAddReplay, error)
	PluginRemove(plugin *svrproto.PluginRemoveArgs) (*svrproto.PluginRemoveReply, error)
	PluginList(plugin *svrproto.PluginListArgs) (*svrproto.PluginListReply, error)
	PluginUpdate(plugin *svrproto.PluginUpdateArgs) (*svrproto.PluginUpdateReply, error)
}

type Provider struct {
	module.ModuleKeeper

	// config
	configList Plugins
	list       Plugins
}

var _ ProviderI = &Provider{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config *config.Plugin) {
	// set plugin list
	for _, item := range config.Lists {
		if err := p.configList.AppendPlugin(moduletypes.Plugin{
			Path:       item.Path,
			Unique:     item.Unique,
			CreateTime: uint64(time.Now().Unix()),
			LoadedTime: 0,
			Params:     item.Params,
		}); err != nil {
			log.Fatal(err)
		}
	}
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
	switch message.Action {
	case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
		for _, item := range p.configList.plugins {
			if err := p.addPlugin(item); err != nil {
				log.WithFields(log.Fields{"unique": item.Unique, "path": item.Path}).Warn("add plugin failed")
			}
		}
	case kpproto.EVENT_MESSAGE_ACTION_PLUGIN_ADD:
		msg := &kpmsg.EventMessagePluginAdd{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		logFields := log.WithFields(log.Fields{"unique": string(msg.Plugin.Unique), "path": string(msg.Plugin.Path)})
		if msg.Error != nil {
			logFields.WithField("error", string(msg.Error)).Warn("add plugin failed")
			break
		}

		// update plugin
		plugin, _, err := p.list.GetPluginByUnique(string(msg.Plugin.Unique))
		if err != nil {
			logFields.Warn(err)
		}
		plugin.LoadedTime = uint64(time.Now().Unix())

		logFields.Info("add plugin success")
	case kpproto.EVENT_MESSAGE_ACTION_PLUGIN_REMOVE:
		msg := &kpmsg.EventMessagePluginRemove{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		logFields := log.WithFields(log.Fields{"unique": string(msg.Plugin.Unique), "path": string(msg.Plugin.Path)})
		if msg.Error != nil {
			logFields.WithField("error", string(msg.Error)).Warn("remove plugin failed")
			break
		}

		if _, err := p.list.RemovePluginByUnique(string(msg.Plugin.Unique)); err != nil {
			log.Fatal(err)
			break
		}

		logFields.Info("remove plugin success")
	case kpproto.EVENT_MESSAGE_ACTION_PLUGIN_UPDATE:
		msg := &kpmsg.EventMessagePluginUpdate{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		logFields := log.WithFields(log.Fields{"unique": string(msg.Plugin.Unique), "path": string(msg.Plugin.Path)})
		if msg.Error != nil {
			logFields.WithField("error", string(msg.Error)).Warn("update plugin failed")
			break
		}

		// update
		plugin, _, err := p.list.GetPluginByUnique(string(msg.Plugin.Unique))
		if err != nil {
			logFields.Warn(err)
		}
		params := make(map[string]string)
		for key, item := range msg.Plugin.Params {
			params[key] = string(item)
		}
		plugin.Params = params

		logFields.Info("update plugin success")
	}
}

func (p *Provider) addPlugin(plugin moduletypes.Plugin) error {
	// validate
	if p.list.Exist(plugin.Unique) {
		return PluginUniqueHasExist
	}
	if !kptypes.FileExists(plugin.Path) {
		return PluginFileNotFound
	}

	// send prompt
	params := map[string][]byte{}
	for k, v := range plugin.Params {
		params[k] = []byte(v)
	}

	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLUGIN_ADD, &kpprompt.EventPromptPluginAdd{
		Plugin: &kpproto.PromptPlugin{
			Path:   []byte(plugin.Path),
			Unique: []byte(plugin.Unique),
			Params: params,
		},
	}); err != nil {
		return err
	}

	// append list
	if err := p.list.AppendPlugin(plugin); err != nil {
		return err
	}

	return nil
}

func (p *Provider) ValidateConfig() error {
	return nil
}
