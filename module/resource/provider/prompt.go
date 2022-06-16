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
	"net/url"
	"os"
	"time"
)

func (p *Provider) ResourceAdd(resource *svrproto.ResourceAddArgs) (*svrproto.ResourceAddReply, error) {
	p.input_mutex.Lock()
	defer p.input_mutex.Unlock()

	// uri scheme parse
	parseUrl, err := url.Parse(resource.Path)
	if err != nil {
		return nil, fmt.Errorf("uri scheme invalid. path: %s", resource.Path)
	}
	if parseUrl.Scheme == "" {
		// determine whether the file exists
		_, err := os.Stat(resource.Path)
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not exists. path: %s", resource.Path)
		}
	}
	if resource.End < resource.Seek {
		return nil, fmt.Errorf("end timestamp can not be less than start timestamp")
	}

	// append to playlist
	if err := p.inputs.AppendResource(moduletypes.Resource{
		Path:       resource.Path,
		Unique:     resource.Unique,
		Seek:       resource.Seek,
		End:        resource.End,
		CreateTime: uint64(time.Now().Unix()),
	}); err != nil {
		return nil, err
	}

	reply := &svrproto.ResourceAddReply{}
	reply.Resource.Unique = resource.Unique
	reply.Resource.Path = resource.Path
	reply.Resource.Seek = resource.Seek
	reply.Resource.End = resource.End

	return reply, nil
}

func (p *Provider) ResourceRemove(resource *svrproto.ResourceRemoveArgs) (*svrproto.ResourceRemoveReply, error) {
	p.input_mutex.Lock()
	defer p.input_mutex.Unlock()

	currentResource, err := p.inputs.GetResourceByIndex(p.currentIndex)
	if err != nil {
		return nil, err
	}

	if resource.Unique == currentResource.Unique {
		return nil, CannotRemoveCurrentResource
	}

	// remove resource
	res, err := p.inputs.RemoveResourceByUnique(resource.Unique)
	if err != nil {
		return nil, err
	}
	p.currentIndex = p.currentIndex - 1

	reply := &svrproto.ResourceRemoveReply{}
	reply.Resource.Path = res.Path
	reply.Resource.Unique = res.Unique
	reply.Resource.CreateTime = res.CreateTime
	return reply, nil
}

func (p *Provider) ResourceList(*svrproto.ResourceListArgs) (*svrproto.ResourceListReply, error) {
	res := []svrproto.Resource{}
	for _, item := range p.inputs.resources[p.currentIndex+1:] {
		res = append(res, svrproto.Resource{
			Path:       item.Path,
			Unique:     item.Unique,
			Seek:       item.Seek,
			End:        item.End,
			CreateTime: item.CreateTime,
			StartTime:  item.StartTime,
			EndTime:    item.EndTime,
		})

	}

	reply := &svrproto.ResourceListReply{}
	reply.Resources = res
	return reply, nil
}

func (p *Provider) ResourceAllList(*svrproto.ResourceAllListArgs) (*svrproto.ResourceAllListReply, error) {
	res := []svrproto.Resource{}
	for _, item := range p.inputs.resources {
		res = append(res, svrproto.Resource{
			Path:       item.Path,
			Unique:     item.Unique,
			Seek:       item.Seek,
			End:        item.End,
			CreateTime: item.CreateTime,
			StartTime:  item.StartTime,
			EndTime:    item.EndTime,
		})

	}

	reply := &svrproto.ResourceAllListReply{}
	reply.Resources = res
	return reply, nil
}

func (p *Provider) CoreResourceList() (*svrproto.ResourceListReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_LIST, &kpprompt.EventPromptResourceList{}); err != nil {
		return nil, err
	}

	resourceListMsg := &msg.EventMessageResourceList{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_LIST, func(msg string) bool {
		types.UnmarshalProtoMessage(msg, resourceListMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(resourceListMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", resourceListMsg.Error)
	}

	reply := &svrproto.ResourceListReply{}
	for _, item := range resourceListMsg.Resources {
		reply.Resources = append(reply.Resources, svrproto.Resource{
			Path:   item.Path,
			Unique: item.Unique,
		})
	}

	return reply, nil
}

func (p *Provider) ResourceCurrent(*svrproto.ResourceCurrentArgs) (*svrproto.ResourceCurrentReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_CURRENT, &kpprompt.EventPromptResourceCurrent{}); err != nil {
		return nil, err
	}

	resourceCurrentMsg := &msg.EventMessageResourceCurrent{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_CURRENT, func(msg string) bool {
		types.UnmarshalProtoMessage(msg, resourceCurrentMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(resourceCurrentMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", resourceCurrentMsg.Error)
	}

	currentRes, err := p.inputs.GetResourceByIndex(p.currentIndex)
	if err != nil {
		return nil, err
	}

	reply := &svrproto.ResourceCurrentReply{
		Resource: svrproto.Resource{
			Path:       resourceCurrentMsg.Resource.Path,
			Seek:       resourceCurrentMsg.Resource.Seek,
			End:        resourceCurrentMsg.Resource.End,
			Unique:     resourceCurrentMsg.Resource.Unique,
			CreateTime: currentRes.CreateTime,
			StartTime:  currentRes.StartTime,
			EndTime:    currentRes.EndTime,
		},
		Duration: resourceCurrentMsg.Duration,
		Seek:     resourceCurrentMsg.Seek,
		HitCache: resourceCurrentMsg.HitCache,
	}
	return reply, nil
}

func (p *Provider) ResourceSeek(args *svrproto.ResourceSeekArgs) (*svrproto.ResourceSeekReply, error) {
	p.input_mutex.Lock()
	defer p.input_mutex.Unlock()

	seekRes, searchIndex, err := p.inputs.GetResourceByUnique(args.Unique)
	if err != nil {
		return nil, err
	}

	p.resetInputs[seekRes.Unique] = seekRes.Seek
	p.currentIndex = searchIndex - 1

	if _, err := p.playProvider.PlaySkip(&svrproto.PlaySkipArgs{}); err != nil {
		return nil, err
	}

	reply := &svrproto.ResourceSeekReply{
		Resource: &svrproto.Resource{
			Path:       seekRes.Path,
			Unique:     seekRes.Unique,
			Seek:       seekRes.Seek,
			End:        seekRes.End,
			CreateTime: seekRes.CreateTime,
			StartTime:  seekRes.StartTime,
			EndTime:    seekRes.EndTime,
		},
	}
	return reply, nil
}
