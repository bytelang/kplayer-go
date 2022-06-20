package rpc

import (
	"context"
	"github.com/bytelang/kplayer/module/plugin/provider"
	"net/http"

	"github.com/bytelang/kplayer/types/server"
)

// Plugin rpc
type Plugin struct {
	pi provider.ProviderI
}

func NewPlugin(pi provider.ProviderI) *Plugin {
	return &Plugin{pi: pi}
}

// List  get plugin list
func (s *Plugin) List(r *http.Request, args *server.PluginListArgs, reply *server.PluginListReply) error {
	listResult, err := s.pi.PluginList(context.TODO(), args)
	if err != nil {
		return err
	}

	reply.Plugins = listResult.Plugins

	return nil
}

// Add add plugin
func (s *Plugin) Add(r *http.Request, args *server.PluginAddArgs, reply *server.PluginAddReplay) error {
	addResult, err := s.pi.PluginAdd(context.TODO(), args)
	if err != nil {
		return err
	}

	reply.Plugin = addResult.Plugin
	return nil
}

// Remove remove plugin
func (s *Plugin) Remove(r *http.Request, args *server.PluginRemoveArgs, reply *server.PluginRemoveReply) error {
	removeResult, err := s.pi.PluginRemove(context.TODO(), args)
	if err != nil {
		return err
	}

	reply.Plugin = removeResult.Plugin
	return nil
}

// Update update plugin params
func (s *Plugin) Update(r *http.Request, args *server.PluginUpdateArgs, reply *server.PluginUpdateReply) error {
	updateResult, err := s.pi.PluginUpdate(context.TODO(), args)
	if err != nil {
		return err
	}

	reply.Plugin = updateResult.Plugin
	return nil
}
