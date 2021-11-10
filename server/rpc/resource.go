package rpc

import (
    "github.com/bytelang/kplayer/module/resource/provider"
    "net/http"

    svrproto "github.com/bytelang/kplayer/types/server"
)

// Resource rpc
type Resource struct {
    pi provider.ProviderI
}

func NewResource(ri provider.ProviderI) *Resource {
    return &Resource{pi: ri}
}

// Add add Resource to core
func (s *Resource) Add(r *http.Request, args *svrproto.ResourceAddArgs, reply *svrproto.ResourceAddReply) (err error) {
    addResult, err := s.pi.ResourceAdd(args)
    if err != nil {
        return err
    }

    reply.Resource = addResult.Resource
    return
}

// Remove remove Resource to core
func (s *Resource) Remove(r *http.Request, args *svrproto.ResourceRemoveArgs, reply *svrproto.ResourceRemoveReply) (err error) {
    removeResult, err := s.pi.ResourceRemove(args)
    if err != nil {
        return err
    }

    reply.Resource = removeResult.Resource
    return
}

// List get untreated resource list
func (s *Resource) List(r *http.Request, args *svrproto.ResourceListArgs, reply *svrproto.ResourceListReply) (err error) {
    listResult, err := s.pi.ResourceList(args)
    if err != nil {
        return err
    }

    reply.Resources = listResult.Resources
    return
}

// AllList get all resource list
func (s *Resource) AllList(r *http.Request, args *svrproto.ResourceAllListArgs, reply *svrproto.ResourceAllListReply) (err error) {
    listResult, err := s.pi.ResourceAllList(args)
    if err != nil {
        return err
    }

    reply.Resources = listResult.Resources
    return
}

// Current get current play resource
func (s *Resource) Current(r *http.Request, args *svrproto.ResourceCurrentArgs, reply *svrproto.ResourceCurrentReply) (err error) {
    currentResource, err := s.pi.ResourceCurrent(args)
    if err != nil {
        return err
    }

    reply.Resource = currentResource.Resource
    return
}
