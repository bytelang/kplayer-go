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
    "os"
    "time"
)

func (p *Provider) ResourceAdd(resource *svrproto.ResourceAddArgs) (*svrproto.ResourceAddReply, error) {
    p.input_mutex.Lock()
    defer p.input_mutex.Unlock()

    // determine whether the file exists
    _, err := os.Stat(resource.Path)
    if os.IsNotExist(err) {
        return nil, fmt.Errorf("file not exists. path: %s", resource.Path)
    }
    if resource.End < resource.Seek {
        return nil, fmt.Errorf("end timestamp can not be less than start timestamp")
    }

    // append to playlist
    p.inputs = append(p.inputs, moduletypes.Resource{
        Path:       resource.Path,
        Unique:     resource.Unique,
        Seek:       resource.Seek,
        End:        resource.End,
        CreateTime: uint64(time.Now().Unix()),
    })
    reply := &svrproto.ResourceAddReply{}
    reply.Resource.Unique = resource.Unique
    reply.Resource.Path = resource.Path

    return reply, nil
}

func (p *Provider) ResourceRemove(resource *svrproto.ResourceRemoveArgs) (*svrproto.ResourceRemoveReply, error) {
    p.input_mutex.Lock()
    defer p.input_mutex.Unlock()

    if resource.Unique == p.inputs[p.currentIndex].Unique {
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
    for _, item := range p.inputs[p.currentIndex+1:] {
        res = append(res, svrproto.Resource{
            Path:       item.Path,
            Unique:     item.Unique,
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
    for _, item := range p.inputs {
        res = append(res, svrproto.Resource{
            Path:       item.Path,
            Unique:     item.Unique,
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
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_LIST, &kpprompt.EventPromptResourceList{
    }); err != nil {
        return nil, err
    }

    resourceListMsg := &msg.EventMessageResourceList{}
    keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_RESOURCE_LIST, func(msg []byte) bool {
        types.UnmarshalProtoMessage(msg, resourceListMsg)
        return true
    })
    defer keeperCtx.Close()

    if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
        return nil, err
    }

    // wait context
    keeperCtx.Wait()
    if resourceListMsg.Error != nil {
        return nil, fmt.Errorf("%s", string(resourceListMsg.Error))
    }

    reply := &svrproto.ResourceListReply{}
    for _, item := range resourceListMsg.Resources {
        reply.Resources = append(reply.Resources, svrproto.Resource{
            Path:   string(item.Path),
            Unique: string(item.Unique),
        })
    }

    return reply, nil
}

func (p *Provider) ResourceCurrent(*svrproto.ResourceCurrentArgs) (*svrproto.ResourceCurrentReply, error) {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_CURRENT, &kpprompt.EventPromptResourceCurrent{
    }); err != nil {
        return nil, err
    }

    resourceCurrentMsg := &msg.EventMessageResourceCurrent{}
    keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_RESOURCE_CURRENT, func(msg []byte) bool {
        types.UnmarshalProtoMessage(msg, resourceCurrentMsg)
        return true
    })
    defer keeperCtx.Close()

    if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
        return nil, err
    }

    // wait context
    keeperCtx.Wait()
    if resourceCurrentMsg.Error != nil {
        return nil, fmt.Errorf("%s", string(resourceCurrentMsg.Error))
    }

    currentRes := p.inputs[p.currentIndex]
    reply := &svrproto.ResourceCurrentReply{
        Resource: svrproto.Resource{
            Path:       string(resourceCurrentMsg.Resource.Path),
            Unique:     string(resourceCurrentMsg.Resource.Unique),
            CreateTime: currentRes.CreateTime,
            StartTime:  currentRes.StartTime,
            EndTime:    currentRes.EndTime,
        },
        Duration: resourceCurrentMsg.Duration,
        Seek:     resourceCurrentMsg.Seek,
    }
    return reply, nil
}

func (p *Provider) ResourceSeek(args *svrproto.ResourceSeekArgs) (*svrproto.ResourceSeekReply, error) {
    p.input_mutex.Lock()
    defer p.input_mutex.Unlock()

    currentRes := p.inputs[p.currentIndex]
    if currentRes.Unique != args.Unique {
        return nil, fmt.Errorf("unique name resource has played. seek unique: %s. current resource unique: %s", args.Unique, currentRes.Unique)
    }

    if _, err := p.playProvider.PlaySkip(&svrproto.PlaySkipArgs{}); err != nil {
        return nil, err
    }

    p.resetInputs[currentRes.Unique] = currentRes.Seek
    p.inputs[p.currentIndex].Seek = args.Seek
    p.currentIndex = p.currentIndex - 1

    reply := &svrproto.ResourceSeekReply{
        Resource: &svrproto.Resource{
            Path:       currentRes.Path,
            Unique:     currentRes.Unique,
            CreateTime: currentRes.CreateTime,
            StartTime:  currentRes.StartTime,
            EndTime:    currentRes.EndTime,
        },
    }
    return reply, nil
}
