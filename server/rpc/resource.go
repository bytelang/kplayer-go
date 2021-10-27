package rpc

import (
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/core/proto/msg"
    "net/http"

    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
    svrproto "github.com/bytelang/kplayer/types/server"
)

const playModuleName = "play"

// Resource rpc
type Resource struct {
    mm module.ModuleManager
}

func NewResource(manager module.ModuleManager) *Resource {
    return &Resource{mm: manager}
}

// Add add Resource to core
func (s *Resource) Add(r *http.Request, args *svrproto.AddResourceArgs, reply *svrproto.AddResourceReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{
        Resource: &kpproto.PromptResource{
            Path:                 []byte(args.Res.Path),
            Unique:               []byte(args.Res.Unique),
            XXX_NoUnkeyedLiteral: struct{}{},
            XXX_unrecognized:     nil,
            XXX_sizecache:        0,
        },
    }); err != nil {
        return err
    }

    resourceModule := s.mm[playModuleName]
    resourceAddMsg := &msg.EventMessageResourceAdd{}

    keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_RESOURCE_ADD, func(msg []byte) bool {
        types.UnmarshalProtoMessage(msg, resourceAddMsg)
        return string(resourceAddMsg.Resource.Unique) == args.Res.Unique
    })
    defer keeperCtx.Close()

    if err := resourceModule.RegisterKeeperChannel(keeperCtx); err != nil {
        return err
    }

    // wait context
    keeperCtx.Wait()

    reply.Res.Unique = string(resourceAddMsg.Resource.Unique)
    reply.Res.Path = string(resourceAddMsg.Resource.Path)

    return nil
}

// Remove remove Resource to core
func (s *Resource) Remove(r *http.Request, args *svrproto.RemoveResourceArgs, reply *svrproto.RemoveResourceReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_REMOVE, &kpprompt.EventPromptResourceRemove{
        Unique: []byte(args.Unique),
    }); err != nil {
        return err
    }

    ResourceModule := s.mm[playModuleName]
    resourceRemoveMsg := &msg.EventMessageResourceRemove{}

    keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_RESOURCE_REMOVE, func(msg []byte) bool {
        types.UnmarshalProtoMessage(msg,resourceRemoveMsg)
        return string(resourceRemoveMsg.Resource.Unique) == args.Unique
    })
    defer keeperCtx.Close()

    if err := ResourceModule.RegisterKeeperChannel(keeperCtx); err != nil {
        return err
    }

    // wait context
    keeperCtx.Wait()

    reply.Res.Unique = string(resourceRemoveMsg.Resource.Unique)
    reply.Res.Path = string(resourceRemoveMsg.Resource.Path)

    return nil
}
