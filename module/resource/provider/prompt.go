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

func (p *Provider) ResourceAdd(resource *svrproto.ResourceAddArgs) (*svrproto.ResourceAddReply, error) {
    p.input_mutex.Lock()
    defer p.input_mutex.Unlock()

    p.inputs = append(p.inputs, moduletypes.Resource{
        Path:       resource.Path,
        Unique:     resource.Unique,
        CreateTime: uint64(time.Now().Unix()),
    })
    reply := &svrproto.ResourceAddReply{}
    reply.Resource.Unique = string(resource.Unique)
    reply.Resource.Path = string(resource.Path)

    return reply, nil
}

func (p *Provider) ResourceRemove(resource *svrproto.ResourceRemoveArgs) (*svrproto.ResourceRemoveReply, error) {
    p.input_mutex.Lock()
    defer p.input_mutex.Unlock()

    reply := &svrproto.ResourceRemoveReply{}
    for _, item := range p.inputs {
        if item.Unique == resource.Unique {
            reply.Resource.Path = item.Path
            reply.Resource.Unique = item.Unique
            reply.Resource.CreateTime = item.CreateTime
            return reply, nil
        }
    }

    return nil, fmt.Errorf("resource not found. unique name: %s", resource.Unique)
}

func (p *Provider) ResourceList(*svrproto.ResourceListArgs) (*svrproto.ResourceListReply, error) {
    res := []svrproto.Resource{}
    for _, item := range p.inputs[p.currentIndex:] {
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
