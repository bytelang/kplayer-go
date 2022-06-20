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
	"path/filepath"
	"strings"
	"time"
)

type ProviderI interface {
	PluginAdd(ctx context.Context, plugin *svrproto.PluginAddArgs) (*svrproto.PluginAddReplay, error)
	PluginRemove(ctx context.Context, plugin *svrproto.PluginRemoveArgs) (*svrproto.PluginRemoveReply, error)
	PluginList(ctx context.Context, plugin *svrproto.PluginListArgs) (*svrproto.PluginListReply, error)
	PluginUpdate(ctx context.Context, plugin *svrproto.PluginUpdateArgs) (*svrproto.PluginUpdateReply, error)
}

type Provider struct {
	module.ModuleKeeper
	svrproto.UnimplementedPluginGreeterServer

	// config
	list Plugins
}

var _ ProviderI = &Provider{}

func NewProvider() *Provider {
	return &Provider{}
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config *config.Plugin) {
	// set plugin list
	for _, item := range config.Lists {
		uniqueName := item.Unique
		if len(uniqueName) == 0 {
			uniqueName = kptypes.GetRandString(6)
		}

		if err := p.list.AppendPlugin(moduletypes.Plugin{
			Path:       GetPluginPath(item.Path),
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
	existName := []string{}

	// init plugin
	for _, item := range p.list.plugins {
		pluginName := strings.TrimSuffix(filepath.Base(item.Path), filepath.Ext(item.Path))
		if pluginName == "" {
			return fmt.Errorf("plugin path cannot be empty")
		}
		if kptypes.ArrayInString(existName, item.Unique) {
			return PluginUniqueHasExist
		}

		logField := log.WithFields(log.Fields{"name": pluginName, "path": item.Path})
		if err := InitPluginFile(pluginName, item.Path); err != nil {
			if _, ok := err.(kptypes.ApiError); ok {
				logField.Error("plugin request information failed")
				return err
			}

			if !kptypes.FileExists(item.Path) {
				logField.Error("plugin initialization failed")
				return err
			}

			logField.Warn(fmt.Sprintf("plugin file exist, but plugin is no registration. %s", err))
		}

		existName = append(existName, item.Unique)
	}

	// init resource
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
			resourceFilePath = filepath.Join("resource/font.ttf")
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
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_ADD:
		msg := &kpmsg.EventMessagePluginAdd{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		logFields := log.WithFields(log.Fields{"unique": msg.Plugin.Unique, "path": msg.Plugin.Path, "author": msg.Plugin.Author})
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
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_REMOVE:
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
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_UPDATE:
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
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
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
				if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLUGIN_ADD, &kpprompt.EventPromptPluginAdd{
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

func (p *Provider) BeginRunning() {
	for _, item := range p.list.plugins {
		// read plugin file
		fileContent, err := kptypes.ReadPlugin(item.Path)
		if err != nil {
			log.WithFields(log.Fields{"path": item.Path, "unique": item.Unique}).Fatal("read plugin file failed")
		}

		if err := core.GetLibKplayerInstance().AddPlugin(&kpprompt.EventPromptPluginAdd{
			Plugin: &kpprompt.PromptPlugin{
				Path:    item.Path,
				Content: fileContent,
				Unique:  item.Unique,
				Params:  item.Params,
			},
		}); err != nil {
			log.WithFields(log.Fields{"unique": item.Unique, "path": item.Path, "error": err}).Trace("add plugin failed")
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

	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLUGIN_ADD, &kpprompt.EventPromptPluginAdd{
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
