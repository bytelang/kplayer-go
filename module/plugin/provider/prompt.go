package provider

import (
	"fmt"
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	"github.com/bytelang/kplayer/types"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/bytelang/kplayer/types/core/proto/msg"
	kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
	moduletypes "github.com/bytelang/kplayer/types/module"
	svrproto "github.com/bytelang/kplayer/types/server"
	"time"
)

func (p *Provider) PluginAdd(args *svrproto.PluginAddArgs) (*svrproto.PluginAddReplay, error) {
	if err := p.addPlugin(moduletypes.Plugin{
		Path:       GetPluginPath(args.Plugin.Path, p.home),
		Unique:     args.Plugin.Unique,
		CreateTime: uint64(time.Now().Unix()),
		Params:     args.Plugin.Params,
	}); err != nil {
		return nil, err
	}

	// wait for message
	pluginAddMsg := &msg.EventMessagePluginAdd{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_PLUGIN_ADD, func(msg []byte) bool {
		types.UnmarshalProtoMessage(msg, pluginAddMsg)
		return types.NewKPString(pluginAddMsg.Plugin.Unique).Equal(args.Plugin.Unique)
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()

	if pluginAddMsg.Error != nil {
		return nil, fmt.Errorf("%s", types.NewKPString(pluginAddMsg.Error))
	}

	// get plugin
	plugin, _, err := p.list.GetPluginByUnique(string(pluginAddMsg.Plugin.Unique))
	if err != nil {
		return nil, err
	}

	reply := &svrproto.PluginAddReplay{}
	reply.Plugin.Path = plugin.Path
	reply.Plugin.Unique = plugin.Unique
	reply.Plugin.Params = plugin.Params
	reply.Plugin.CreateTime = plugin.CreateTime
	reply.Plugin.LoadedTime = plugin.LoadedTime

	return reply, nil
}

func (p *Provider) PluginRemove(args *svrproto.PluginRemoveArgs) (*svrproto.PluginRemoveReply, error) {
	// validate
	if !p.list.Exist(args.Unique) {
		return nil, PluginUniqueNotFound
	}

	// send prompt
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLUGIN_REMOVE, &kpprompt.EventPromptPluginRemove{
		Unique: []byte(args.Unique),
	}); err != nil {
		return nil, err
	}

	pluginRemoveMsg := &msg.EventMessagePluginRemove{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_PLUGIN_REMOVE, func(msg []byte) bool {
		types.UnmarshalProtoMessage(msg, pluginRemoveMsg)
		return types.NewKPString(pluginRemoveMsg.Plugin.Unique).Equal(args.Unique)
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if pluginRemoveMsg.Error != nil {
		return nil, fmt.Errorf("%s", string(pluginRemoveMsg.Error))
	}

	reply := &svrproto.PluginRemoveReply{}
	reply.Plugin.Path = string(pluginRemoveMsg.Plugin.Path)
	reply.Plugin.Unique = string(pluginRemoveMsg.Plugin.Unique)
	reply.Plugin.Params = make(map[string]string)
	for k, v := range pluginRemoveMsg.Plugin.Params {
		reply.Plugin.Params[k] = string(v)
	}

	return reply, nil
}

func (p *Provider) PluginList(plugin *svrproto.PluginListArgs) (*svrproto.PluginListReply, error) {
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

func (p *Provider) PluginListFromCore(args *svrproto.PluginListArgs) (*svrproto.PluginListReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLUGIN_LIST, &kpprompt.EventPromptPluginList{}); err != nil {
		return nil, err
	}

	pluginListMsg := &msg.EventMessagePluginList{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_PLUGIN_LIST, func(msg []byte) bool {
		types.UnmarshalProtoMessage(msg, pluginListMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if pluginListMsg.Error != nil {
		return nil, fmt.Errorf("%s", string(pluginListMsg.Error))
	}

	reply := &svrproto.PluginListReply{}
	for _, item := range pluginListMsg.Plugins {
		params := map[string]string{}
		for k, v := range item.Params {
			params[k] = string(v)
		}

		reply.Plugins = append(reply.Plugins, &svrproto.Plugin{
			Path:   string(item.Path),
			Unique: string(item.Unique),
			Params: params,
		})
	}

	return reply, nil
}

func (p *Provider) PluginUpdate(args *svrproto.PluginUpdateArgs) (*svrproto.PluginUpdateReply, error) {
	// validate
	if !p.list.Exist(args.Unique) {
		return nil, PluginUniqueNotFound
	}

	// send prompt
	coreKplayer := core.GetLibKplayerInstance()

	argParams := map[string][]byte{}
	for k, v := range args.Params {
		argParams[k] = []byte(v)
	}
	if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLUGIN_UPDATE, &kpprompt.EventPromptPluginUpdate{
		Unique: []byte(args.Unique),
		Params: argParams,
	}); err != nil {
		return nil, err
	}

	pluginUpdateMsg := &msg.EventMessagePluginUpdate{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_PLUGIN_UPDATE, func(msg []byte) bool {
		types.UnmarshalProtoMessage(msg, pluginUpdateMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if pluginUpdateMsg.Error != nil {
		return nil, fmt.Errorf("%s", string(pluginUpdateMsg.Error))
	}

	replyParams := map[string]string{}
	for k, v := range pluginUpdateMsg.Plugin.Params {
		replyParams[k] = string(v)
	}

	return &svrproto.PluginUpdateReply{
		Plugin: &svrproto.Plugin{
			Path:   string(pluginUpdateMsg.Plugin.Path),
			Unique: string(pluginUpdateMsg.Plugin.Unique),
			Params: replyParams,
		},
	}, nil
}
