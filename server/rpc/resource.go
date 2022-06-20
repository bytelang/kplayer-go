package rpc

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/module/resource/provider"
	svrproto "github.com/bytelang/kplayer/types/server"
	"net/http"
	"time"
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
	addResult, err := s.pi.ResourceAdd(context.TODO(), args)
	if err != nil {
		return err
	}

	reply.Resource = addResult.Resource
	return
}

func (s *Resource) Seek(r *http.Request, args *svrproto.ResourceSeekArgs, reply *svrproto.ResourceSeekReply) (err error) {
	seekResult, err := s.pi.ResourceSeek(context.TODO(), args)
	if err != nil {
		return err
	}
	reply.Resource = seekResult.Resource
	return
}

// Remove remove Resource to core
func (s *Resource) Remove(r *http.Request, args *svrproto.ResourceRemoveArgs, reply *svrproto.ResourceRemoveReply) (err error) {
	removeResult, err := s.pi.ResourceRemove(context.TODO(), args)
	if err != nil {
		return err
	}

	reply.Resource = removeResult.Resource
	return
}

// List get untreated resource list
func (s *Resource) List(r *http.Request, args *svrproto.ResourceListArgs, reply *svrproto.ResourceListReply) (err error) {
	listResult, err := s.pi.ResourceList(context.TODO(), args)
	if err != nil {
		return err
	}

	reply.Resources = listResult.Resources
	return
}

// AllList get all resource list
func (s *Resource) AllList(r *http.Request, args *svrproto.ResourceListAllArgs, reply *svrproto.ResourceListAllReply) (err error) {
	listResult, err := s.pi.ResourceListAll(context.TODO(), args)
	if err != nil {
		return err
	}

	reply.Resources = listResult.Resources
	return
}

// Current get current play resource
func (s *Resource) Current(r *http.Request, args *svrproto.ResourceCurrentArgs, reply *svrproto.ResourceCurrentReply) (err error) {
	currentResource, err := s.pi.ResourceCurrent(context.TODO(), args)
	if err != nil {
		return err
	}

	(*reply) = *currentResource

	resourceDuration := time.Duration(time.Second * time.Duration(currentResource.Duration))
	reply.DurationFormat = fmt.Sprintf("%d:%d:%d", uint64(resourceDuration.Hours()), uint64(resourceDuration.Minutes())%60, uint64(resourceDuration.Seconds())%60)

	resourceSeek := time.Duration(time.Second * time.Duration(currentResource.Seek))
	reply.SeekFormat = fmt.Sprintf("%d:%d:%d", uint64(resourceSeek.Hours()), uint64(resourceSeek.Minutes())%60, uint64(resourceSeek.Seconds())%60)
	return
}
