package rpc

import (
    "fmt"
    "github.com/bytelang/kplayer/core"
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/types"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    "github.com/bytelang/kplayer/types/core/proto/msg"
    kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
    svrproto "github.com/bytelang/kplayer/types/server"
    "net/http"

    "github.com/bytelang/kplayer/types/server"
)

const pluginModuleName = "plugin"

// Plugin rpc
type Plugin struct {
    mm module.ModuleManager
}

func NewPlugin(manager module.ModuleManager) *Plugin {
    return &Plugin{mm: manager}
}

// List  get plugin list
func (s *Plugin) List(r *http.Request, args *server.PluginListArgs, reply *server.PluginListReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLUGIN_LIST, &kpprompt.EventPromptPluginList{
    }); err != nil {
        return err
    }

    pluginListMsg := &msg.EventMessagePluginList{}
    keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_PLUGIN_LIST, func(msg []byte) bool {
        types.UnmarshalProtoMessage(msg, pluginListMsg)
        return true
    })
    defer keeperCtx.Close()

    pluginModule := s.mm[pluginModuleName]
    if err := pluginModule.RegisterKeeperChannel(keeperCtx); err != nil {
        return err
    }

    // wait context
    keeperCtx.Wait()

    reply.Plugins = make([]svrproto.Plugin, 0)
    for _, item := range pluginListMsg.Plugins {
        reply.Plugins = append(reply.Plugins, svrproto.Plugin{
            Path:   string(item.Path),
            Unique: string(item.Unique),
        })
    }

    return nil
}

// Add add plugin
func (s *Plugin) Add(r *http.Request, args *server.PluginAddArgs, reply *server.PluginAddReplay) error {
    params := map[string][]byte{}
    for k, v := range args.Plugin.Params {
        params[k] = []byte(v)
    }

    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLUGIN_ADD, &kpprompt.EventPromptPluginAdd{
        Plugin: &kpproto.PromptPlugin{
            Path:   []byte(args.Plugin.Path),
            Unique: []byte(args.Plugin.Unique),
            Params: params,
        },
    }); err != nil {
        return err
    }

    pluginAddMsg := &msg.EventMessagePluginAdd{}
    keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_PLUGIN_ADD, func(msg []byte) bool {
        types.UnmarshalProtoMessage(msg, pluginAddMsg)
        return types.NewKPString(pluginAddMsg.Plugin.Path).Equal(args.Plugin.Path) && types.NewKPString(pluginAddMsg.Plugin.Unique).Equal(args.Plugin.Unique)
    })
    defer keeperCtx.Close()

    pluginModule := s.mm[pluginModuleName]
    if err := pluginModule.RegisterKeeperChannel(keeperCtx); err != nil {
        return err
    }

    // wait context
    keeperCtx.Wait()

    if pluginAddMsg.Error != nil {
        return fmt.Errorf("%s", types.NewKPString(pluginAddMsg.Error))
    }

    reply.Plugin.Path = types.NewKPString(pluginAddMsg.Plugin.Path).String()
    reply.Plugin.Unique = types.NewKPString(pluginAddMsg.Plugin.Unique).String()

    return nil
}
