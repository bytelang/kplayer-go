package provider

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/bytelang/kplayer/types/core/proto/msg"
	kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
	moduletypes "github.com/bytelang/kplayer/types/module"
	svrproto "github.com/bytelang/kplayer/types/server"
	log "github.com/sirupsen/logrus"
	"path/filepath"
	"strings"
	"time"
)

func (p *Provider) PluginAdd(ctx context.Context, args *svrproto.PluginAddArgs) (*svrproto.PluginAddReplay, error) {
	// download plugin file
	pluginName := strings.TrimSuffix(filepath.Base(args.Path), filepath.Ext(args.Path))
	if pluginName == "" {
		return nil, fmt.Errorf("plugin path cannot be empty")
	}
	logField := log.WithFields(log.Fields{"name": pluginName, "path": args.Path})
	if err := InitPluginFile(pluginName, GetPluginPath(args.Path)); err != nil {
		if _, ok := err.(kptypes.ApiError); ok {
			logField.Error("plugin request information failed")
			return nil, err
		}

		if !kptypes.FileExists(args.Path) {
			logField.Error("plugin initialization failed")
			return nil, err
		}

		logField.Warn(fmt.Sprintf("plugin file exist, but plugin is no registration. %s", err))
	}

	// add plugin prompt
	if err := p.addPlugin(moduletypes.Plugin{
		Path:       GetPluginPath(args.Path),
		Unique:     args.Unique,
		CreateTime: uint64(time.Now().Unix()),
		Params:     args.Params,
	}); err != nil {
		return nil, err
	}

	// get plugin
	plugin, _, err := p.list.GetPluginByUnique(args.Unique)
	if err != nil {
		return nil, err
	}

	reply := &svrproto.PluginAddReplay{Plugin: &svrproto.Plugin{}}
	reply.Plugin.Path = plugin.Path
	reply.Plugin.Unique = plugin.Unique
	reply.Plugin.Params = plugin.Params
	reply.Plugin.CreateTime = plugin.CreateTime
	reply.Plugin.LoadedTime = plugin.LoadedTime

	return reply, nil
}

func (p *Provider) PluginRemove(ctx context.Context, args *svrproto.PluginRemoveArgs) (*svrproto.PluginRemoveReply, error) {
	// validate
	if !p.list.Exist(args.Unique) {
		return nil, PluginUniqueNotFound
	}

	// send prompt
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLUGIN_REMOVE, &kpprompt.EventPromptPluginRemove{
		Unique: args.Unique,
	}); err != nil {
		return nil, err
	}

	pluginRemoveMsg := &msg.EventMessagePluginRemove{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_REMOVE, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, pluginRemoveMsg)
		return pluginRemoveMsg.Plugin.Unique == args.Unique
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(pluginRemoveMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", pluginRemoveMsg.Error)
	}

	reply := &svrproto.PluginRemoveReply{Plugin: &svrproto.Plugin{}}
	reply.Plugin.Path = pluginRemoveMsg.Plugin.Path
	reply.Plugin.Unique = pluginRemoveMsg.Plugin.Unique
	reply.Plugin.Params = make(map[string]string)
	for k, v := range pluginRemoveMsg.Plugin.Params {
		reply.Plugin.Params[k] = v
	}

	return reply, nil
}

func (p *Provider) PluginList(ctx context.Context, args *svrproto.PluginListArgs) (*svrproto.PluginListReply, error) {
	reply := &svrproto.PluginListReply{}
	for _, item := range p.list.plugins {
		reply.Plugins = append(reply.Plugins, &svrproto.Plugin{
			Path:       item.Path,
			Unique:     item.Unique,
			CreateTime: item.CreateTime,
			LoadedTime: item.LoadedTime,
			Params:     item.Params,
		})
	}

	return reply, nil
}

func (p *Provider) PluginListFromCore(ctx context.Context, args *svrproto.PluginListArgs) (*svrproto.PluginListReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLUGIN_LIST, &kpprompt.EventPromptPluginList{}); err != nil {
		return nil, err
	}

	pluginListMsg := &msg.EventMessagePluginList{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_LIST, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, pluginListMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(pluginListMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", pluginListMsg.Error)
	}

	reply := &svrproto.PluginListReply{}
	for _, item := range pluginListMsg.Plugins {
		params := map[string]string{}
		for k, v := range item.Params {
			params[k] = v
		}

		reply.Plugins = append(reply.Plugins, &svrproto.Plugin{
			Path:   item.Path,
			Unique: item.Unique,
			Params: params,
		})
	}

	return reply, nil
}

func (p *Provider) PluginUpdate(ctx context.Context, args *svrproto.PluginUpdateArgs) (*svrproto.PluginUpdateReply, error) {
	// validate
	if !p.list.Exist(args.Unique) {
		return nil, PluginUniqueNotFound
	}

	// send prompt
	coreKplayer := core.GetLibKplayerInstance()

	argParams := map[string]string{}
	for k, v := range args.Params {
		argParams[k] = v
	}
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLUGIN_UPDATE, &kpprompt.EventPromptPluginUpdate{
		Unique: args.Unique,
		Params: argParams,
	}); err != nil {
		return nil, err
	}

	pluginUpdateMsg := &msg.EventMessagePluginUpdate{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_UPDATE, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, pluginUpdateMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(pluginUpdateMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", pluginUpdateMsg.Error)
	}

	replyParams := map[string]string{}
	for k, v := range pluginUpdateMsg.Plugin.Params {
		replyParams[k] = v
	}

	return &svrproto.PluginUpdateReply{
		Plugin: &svrproto.Plugin{
			Path:   pluginUpdateMsg.Plugin.Path,
			Unique: pluginUpdateMsg.Plugin.Unique,
			Params: replyParams,
		},
	}, nil
}
