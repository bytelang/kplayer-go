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
	"net/url"
	"os"
	"time"
)

func (p *Provider) ResourceAdd(ctx context.Context, args *svrproto.ResourceAddArgs) (*svrproto.ResourceAddReply, error) {
	p.input_mutex.Lock()
	defer p.input_mutex.Unlock()

	// uri scheme parse
	if !args.MixResourceType {
		parseUrl, err := url.Parse(args.Path)
		if err != nil {
			return nil, fmt.Errorf("uri scheme invalid. path: %s", args.Path)
		}
		if parseUrl.Scheme == "" {
			// determine whether the file exists
			_, err := os.Stat(args.Path)
			if os.IsNotExist(err) {
				return nil, fmt.Errorf("file not exists. path: %s", args.Path)
			}
		}
	} else {
		for _, item := range args.Groups {
			parseUrl, err := url.Parse(item.Path)
			if err != nil {
				return nil, fmt.Errorf("media_type %s. uri scheme invalid. path: %s", item.MediaType, item.Path)
			}
			if parseUrl.Scheme == "" {
				// determine whether the file exists
				_, err := os.Stat(item.Path)
				if os.IsNotExist(err) {
					return nil, fmt.Errorf("media_type: %s. file not exists. path: %s", item.MediaType, item.Path)
				}
			}
		}
	}

	if args.End < args.Seek {
		return nil, fmt.Errorf("end timestamp can not be less than start timestamp")
	}

	// append to playlist
	primaryPath := args.Path
	moduleGroups := TransferServerToModuleResourceGroup(args.Groups)
	if args.MixResourceType {
		_, _, primaryResourceGroup := CalcMixResourceGroupPrimaryPath(moduleGroups)
		primaryPath = primaryResourceGroup.Path
	}
	moduleResource := moduletypes.Resource{
		Path:            primaryPath,
		Unique:          args.Unique,
		Seek:            args.Seek,
		End:             args.End,
		MixResourceType: args.MixResourceType,
		CreateTime:      uint64(time.Now().Unix()),
		Groups:          moduleGroups,
	}

	if err := p.inputs.AppendResource(moduleResource); err != nil {
		return nil, err
	}

	reply := &svrproto.ResourceAddReply{Resource: &svrproto.Resource{}}
	reply.Resource.Unique = moduleResource.Unique
	reply.Resource.Path = moduleResource.Path
	reply.Resource.Seek = moduleResource.Seek
	reply.Resource.End = moduleResource.End
	reply.Resource.MixResourceType = moduleResource.MixResourceType
	reply.Resource.Groups = args.Groups

	return reply, nil
}

func (p *Provider) ResourceRemove(ctx context.Context, args *svrproto.ResourceRemoveArgs) (*svrproto.ResourceRemoveReply, error) {
	p.input_mutex.Lock()
	defer p.input_mutex.Unlock()

	currentResource, err := p.inputs.GetResourceByIndex(p.currentIndex)
	if err != nil {
		return nil, err
	}

	if args.Unique == currentResource.Unique {
		return nil, CannotRemoveCurrentResource
	}

	// remove resource
	res, index, err := p.inputs.RemoveResourceByUnique(args.Unique)
	if err != nil {
		return nil, err
	}
	if index < p.currentIndex {
		p.currentIndex = p.currentIndex - 1
	}

	reply := &svrproto.ResourceRemoveReply{Resource: &svrproto.ResourceRemoveReply_Resource{}}
	reply.Resource.Path = res.Path
	reply.Resource.Unique = res.Unique
	reply.Resource.CreateTime = res.CreateTime
	return reply, nil
}

func (p *Provider) ResourceList(ctx context.Context, args *svrproto.ResourceListArgs) (*svrproto.ResourceListReply, error) {
	var res []*svrproto.Resource
	for _, item := range p.inputs.resources[p.currentIndex+1:] {
		var groups []*svrproto.MixResourceGroup
		for _, groupItem := range item.Groups {
			groups = append(groups, &svrproto.MixResourceGroup{
				Path: groupItem.Path,
				MediaType: func() svrproto.ResourceMediaType {
					if groupItem.MediaType == moduletypes.ResourceMediaType_audio {
						return svrproto.ResourceMediaType_audio
					}
					return svrproto.ResourceMediaType_video
				}(),
				PersistentLoop: groupItem.PersistentLoop,
			})
		}

		res = append(res, &svrproto.Resource{
			Path:            item.Path,
			Unique:          item.Unique,
			Seek:            item.Seek,
			End:             item.End,
			CreateTime:      item.CreateTime,
			StartTime:       item.StartTime,
			EndTime:         item.EndTime,
			MixResourceType: item.MixResourceType,
			Groups:          groups,
		})

	}

	reply := &svrproto.ResourceListReply{}
	reply.Resources = res
	return reply, nil
}

func (p *Provider) ResourceListAll(ctx context.Context, args *svrproto.ResourceListAllArgs) (*svrproto.ResourceListAllReply, error) {
	var res []*svrproto.Resource
	for _, item := range p.inputs.resources {
		var groups []*svrproto.MixResourceGroup
		for _, groupItem := range item.Groups {
			groups = append(groups, &svrproto.MixResourceGroup{
				Path: groupItem.Path,
				MediaType: func() svrproto.ResourceMediaType {
					if groupItem.MediaType == moduletypes.ResourceMediaType_audio {
						return svrproto.ResourceMediaType_audio
					}
					return svrproto.ResourceMediaType_video
				}(),
				PersistentLoop: groupItem.PersistentLoop,
			})
		}

		res = append(res, &svrproto.Resource{
			Path:            item.Path,
			Unique:          item.Unique,
			Seek:            item.Seek,
			End:             item.End,
			CreateTime:      item.CreateTime,
			StartTime:       item.StartTime,
			EndTime:         item.EndTime,
			MixResourceType: item.MixResourceType,
			Groups:          groups,
		})

	}

	reply := &svrproto.ResourceListAllReply{}
	reply.Resources = res
	return reply, nil
}

func (p *Provider) CoreResourceList() (*svrproto.ResourceListReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_LIST, &kpprompt.EventPromptResourceList{}); err != nil {
		return nil, err
	}

	resourceListMsg := &msg.EventMessageResourceList{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_LIST, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, resourceListMsg)
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
		var groups []*svrproto.MixResourceGroup
		for _, groupItem := range item.Groups {
			groups = append(groups, &svrproto.MixResourceGroup{
				Path: groupItem.Path,
				MediaType: func() svrproto.ResourceMediaType {
					if groupItem.MediaType == kpproto.ResourceMediaType_RESOURCE_MEDIA_TYPE_AUDIO {
						return svrproto.ResourceMediaType_audio
					}
					return svrproto.ResourceMediaType_video
				}(),
				PersistentLoop: groupItem.PersistentLoop,
			})
		}

		mixResourceType := false
		if item.InputType == kpproto.ResourceInputType_RESOURCE_INPUT_TYPE_MIX {
			mixResourceType = true
		}

		reply.Resources = append(reply.Resources, &svrproto.Resource{
			Path:            item.Path,
			Unique:          item.Unique,
			Seek:            item.Seek,
			End:             item.End,
			MixResourceType: mixResourceType,
			Groups:          groups,
		})
	}

	return reply, nil
}

func (p *Provider) ResourceCurrent(ctx context.Context, args *svrproto.ResourceCurrentArgs) (*svrproto.ResourceCurrentReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_CURRENT, &kpprompt.EventPromptResourceCurrent{}); err != nil {
		return nil, err
	}

	resourceCurrentMsg := &msg.EventMessageResourceCurrent{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_CURRENT, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, resourceCurrentMsg)
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

	resourceDuration := time.Second * time.Duration(resourceCurrentMsg.Duration)
	resourceSeek := time.Second * time.Duration(resourceCurrentMsg.Seek)

	// groups
	var groups []*svrproto.MixResourceGroup
	for _, groupItem := range currentRes.Groups {
		groups = append(groups, &svrproto.MixResourceGroup{
			Path: groupItem.Path,
			MediaType: func() svrproto.ResourceMediaType {
				if groupItem.MediaType == moduletypes.ResourceMediaType_audio {
					return svrproto.ResourceMediaType_audio
				}
				return svrproto.ResourceMediaType_video
			}(),
			PersistentLoop: groupItem.PersistentLoop,
		})
	}

	reply := &svrproto.ResourceCurrentReply{
		Resource: &svrproto.Resource{
			Path:            resourceCurrentMsg.Resource.Path,
			Seek:            resourceCurrentMsg.Resource.Seek,
			End:             resourceCurrentMsg.Resource.End,
			Unique:          resourceCurrentMsg.Resource.Unique,
			CreateTime:      currentRes.CreateTime,
			StartTime:       currentRes.StartTime,
			EndTime:         currentRes.EndTime,
			MixResourceType: currentRes.MixResourceType,
			Groups:          groups,
		},
		Duration:       resourceCurrentMsg.Duration,
		DurationFormat: fmt.Sprintf("%d:%d:%d", uint64(resourceDuration.Hours()), uint64(resourceDuration.Minutes())%60, uint64(resourceDuration.Seconds())%60),
		Seek:           resourceCurrentMsg.Seek,
		SeekFormat:     fmt.Sprintf("%d:%d:%d", uint64(resourceSeek.Hours()), uint64(resourceSeek.Minutes())%60, uint64(resourceSeek.Seconds())%60),
		HitCache:       resourceCurrentMsg.HitCache,
	}
	return reply, nil
}

func (p *Provider) ResourceSeek(ctx context.Context, args *svrproto.ResourceSeekArgs) (*svrproto.ResourceSeekReply, error) {
	p.input_mutex.Lock()
	defer p.input_mutex.Unlock()

	var seekRes *moduletypes.Resource
	var seekIndex int
	var err error
	if len(args.Unique) != 0 {
		seekRes, seekIndex, err = p.inputs.GetResourceByUnique(args.Unique)
		if err != nil {
			return nil, err
		}
	} else {
		seekRes, err = p.inputs.GetResourceByIndex(p.currentIndex)
		if err != nil {
			return nil, err
		}
	}

	// send prompt
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_SEEK, &kpprompt.EventPromptResourceSeek{
		Resource: &kpproto.PromptResource{
			Path:   seekRes.Path,
			Unique: seekRes.Unique,
			Seek:   args.Seek,
			End:    -1,
		},
	}); err != nil {
		return nil, err
	}

	resourceSeek := &msg.EventMessageResourceSeek{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_SEEK, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, resourceSeek)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(resourceSeek.Error) != 0 {
		return nil, fmt.Errorf("%s", resourceSeek.Error)
	}

	// groups
	var groups []*svrproto.MixResourceGroup
	for _, groupItem := range seekRes.Groups {
		groups = append(groups, &svrproto.MixResourceGroup{
			Path: groupItem.Path,
			MediaType: func() svrproto.ResourceMediaType {
				if groupItem.MediaType == moduletypes.ResourceMediaType_audio {
					return svrproto.ResourceMediaType_audio
				}
				return svrproto.ResourceMediaType_video
			}(),
			PersistentLoop: groupItem.PersistentLoop,
		})
	}

	reply := &svrproto.ResourceSeekReply{
		Resource: &svrproto.Resource{
			Path:            resourceSeek.Resource.Path,
			Unique:          resourceSeek.Resource.Unique,
			Seek:            resourceSeek.Resource.Seek,
			End:             resourceSeek.Resource.End,
			CreateTime:      seekRes.CreateTime,
			StartTime:       seekRes.StartTime,
			EndTime:         seekRes.EndTime,
			MixResourceType: seekRes.MixResourceType,
			Groups:          groups,
		},
	}

	// update current resource index
	p.currentIndex = seekIndex
	return reply, nil
}
