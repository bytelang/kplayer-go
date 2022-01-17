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
	"path/filepath"
	"strings"
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

	// flag
	home string

	// config
	configList Plugins
	list       Plugins
}

var _ ProviderI = &Provider{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config *config.Plugin, homePath string) {
	p.home = homePath

	// set plugin list
	for _, item := range config.Lists {
		uniqueName := item.Unique
		if len(uniqueName) == 0 {
			uniqueName = kptypes.GetRandString(6)
		}

		if err := p.configList.AppendPlugin(moduletypes.Plugin{
			Path:       GetPluginPath(item.Path, homePath),
			Unique:     uniqueName,
			CreateTime: uint64(time.Now().Unix()),
			LoadedTime: 0,
			Params:     item.Params,
		}); err != nil {
			log.Fatal(err)
		}
	}
}

func (p *Provider) ValidateConfig() error {
	// get version
	kptypes.GetCorePluginVersion()

	// init plugin
	if len(p.configList.plugins) > 0 {
		if err := kptypes.MkDir(filepath.Join(p.home, "plugin")); err != nil {
			return err
		}
	}

	for _, item := range p.configList.plugins {
		pluginName := strings.TrimSuffix(filepath.Base(item.Path), filepath.Ext(item.Path))

		logField := log.WithFields(log.Fields{"name": pluginName, "path": item.Path})
		if err := InitPluginFile(pluginName, item.Path); err != nil {
			if !kptypes.FileExists(item.Path) {
				logField.Error("plugin initialization failed")
				return err
			}

			logField.Warn("plugin file exist, but plugin is no registration.")
		}
	}

	// init resource
	if len(p.configList.plugins) > 0 {
		if err := kptypes.MkDir(filepath.Join(p.home, "resource")); err != nil {
			return err
		}
	}
	resources := []map[string]string{
		{
			"type": "font",
			"name": "default",
		},
	}
	for _, item := range resources {
		logField := log.WithFields(log.Fields{"type": item["type"], "name": item["name"]})

		var resourceFilePath string
		if item["type"] == "font" {
			resourceFilePath = filepath.Join(p.home, "resource/font.ttf")
		}

		if !kptypes.FileExists(resourceFilePath) {
			if err := InitResourceFile(item["type"], item["name"], resourceFilePath); err != nil {
				return err
			}
			logField.Info("resource initialization success")
		}
	}

	return nil
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
	switch message.Action {
	case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
		for _, item := range p.configList.plugins {
			if err := p.addPlugin(item); err != nil {
				log.WithFields(log.Fields{"unique": item.Unique, "path": item.Path, "error": err}).Warn("add plugin failed")
			}
		}
	case kpproto.EVENT_MESSAGE_ACTION_PLUGIN_ADD:
		msg := &kpmsg.EventMessagePluginAdd{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		logFields := log.WithFields(log.Fields{"unique": msg.Plugin.Unique, "path": msg.Plugin.Path})
		if len(msg.Error) != 0 {
			logFields.WithField("error", msg.Error).Warn("add plugin failed")
			break
		}

		// update plugin
		plugin, _, err := p.list.GetPluginByUnique(msg.Plugin.Unique)
		if err != nil {
			logFields.Warn(err)
		}
		plugin.LoadedTime = uint64(time.Now().Unix())

		logFields.Info("add plugin success")
	case kpproto.EVENT_MESSAGE_ACTION_PLUGIN_REMOVE:
		msg := &kpmsg.EventMessagePluginRemove{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		logFields := log.WithFields(log.Fields{"unique": msg.Plugin.Unique, "path": msg.Plugin.Path})
		if len(msg.Error) != 0 {
			logFields.WithField("error", msg.Error).Warn("remove plugin failed")
			break
		}

		if _, err := p.list.RemovePluginByUnique(msg.Plugin.Unique); err != nil {
			log.Fatal(err)
			break
		}

		logFields.Info("remove plugin success")
	case kpproto.EVENT_MESSAGE_ACTION_PLUGIN_UPDATE:
		msg := &kpmsg.EventMessagePluginUpdate{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		logFields := log.WithFields(log.Fields{"unique": msg.Plugin.Unique, "path": msg.Plugin.Path})
		if len(msg.Error) != 0 {
			logFields.WithField("error", msg.Error).Warn("update plugin failed")
			break
		}

		// update
		plugin, _, err := p.list.GetPluginByUnique(msg.Plugin.Unique)
		if err != nil {
			logFields.Warn(err)
		}
		params := make(map[string]string)
		for key, item := range msg.Plugin.Params {
			params[key] = item
		}
		plugin.Params = params

		logFields.Info("update plugin success")
	case kpproto.EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
		// reload failed plugin
		p.list.lock.Lock()
		defer p.list.lock.Unlock()

		for _, item := range p.list.plugins {
			if item.LoadedTime == 0 {
				// send prompt
				params := map[string]string{}
				for k, v := range item.Params {
					params[k] = v
				}

				coreKplayer := core.GetLibKplayerInstance()
				if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLUGIN_ADD, &kpprompt.EventPromptPluginAdd{
					Plugin: &kpprompt.PromptPlugin{
						Path:   item.Path,
						Unique: item.Unique,
						Params: params,
					},
				}); err != nil {
					log.WithFields(log.Fields{"path": item.Path, "unique": item.Unique, "error": err}).Warn("reload plugin failed")
				}
			}
		}
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
	params := map[string]string{}
	for k, v := range plugin.Params {
		params[k] = v
	}

	coreKplayer := core.GetLibKplayerInstance()

	// read plugin file
	fileContent, err := kptypes.ReadPlugin(plugin.Path)
	if err != nil {
		log.WithFields(log.Fields{"path": plugin.Path, "unique": plugin.Unique}).Error("read plugin file failed")
		return err
	}

	if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLUGIN_ADD, &kpprompt.EventPromptPluginAdd{
		Plugin: &kpprompt.PromptPlugin{
			Path:    plugin.Path,
			Content: fileContent,
			Unique:  plugin.Unique,
			Params:  params,
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
