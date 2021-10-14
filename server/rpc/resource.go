package rpc

import (
    "fmt"
    "github.com/bytelang/kplayer/module"
    resourcetype "github.com/bytelang/kplayer/module/resource/types"
    "github.com/bytelang/kplayer/proto/msg"
    "github.com/google/uuid"
    "net/http"

    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/proto"
    prompt "github.com/bytelang/kplayer/proto/prompt"
    "github.com/bytelang/kplayer/server/proto"
)

// Resource rpc
type Resource struct {
    mm module.ModuleManager
}

func NewResource(manager module.ModuleManager) *Resource {
    return &Resource{mm: manager}
}

// Add add Resource to core
func (s *Resource) Add(r *http.Request, args *proto.AddResourceArgs, reply *proto.AddResourceReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_ADD, &prompt.EventPromptResourceAdd{
        Path:   []byte(args.Res.Path),
        Unique: []byte(args.Res.Unique),
    }); err != nil {
        return err
    }

    resourceModule := s.mm[resourcetype.ModuleName]
    keeperCtx := module.NewKeeperContext(uuid.New().String(), kpproto.EVENT_MESSAGE_ACTION_RESOURCE_ADD)
    defer keeperCtx.Close()

    if err := resourceModule.RegisterKeeperChannel(keeperCtx); err != nil {
        return err
    }

    // wait context
    ResourceAddMsg := &msg.EventMessageResourceAdd{}
    if err := keeperCtx.Wait(ResourceAddMsg); err != nil {
        return fmt.Errorf("messge type invalid")
    }

    reply.Res = proto.Resource{
        Path:   string(ResourceAddMsg.Path),
        Unique: string(ResourceAddMsg.Unique),
    }

    return nil
}

// Remove remove Resource to core
func (s *Resource) Remove(r *http.Request, args *proto.RemoveResourceArgs, reply *proto.RemoveResourceReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_REMOVE, &prompt.EventPromptResourceRemove{
        Unique: []byte(args.Unique),
    }); err != nil {
        return err
    }

    ResourceModule := s.mm[resourcetype.ModuleName]

    keeperCtx := module.NewKeeperContext(uuid.New().String(), kpproto.EVENT_MESSAGE_ACTION_RESOURCE_REMOVE)
    defer keeperCtx.Close()

    if err := ResourceModule.RegisterKeeperChannel(keeperCtx); err != nil {
        return err
    }

    // wait context
    ResourceRemoveMsg := &msg.EventMessageResourceRemove{}
    if err := keeperCtx.Wait(ResourceRemoveMsg); err != nil {
        return fmt.Errorf("messge type invalid")
    }

    reply.Res = proto.Resource{
        Path:   string(ResourceRemoveMsg.Path),
        Unique: string(ResourceRemoveMsg.Unique),
    }

    return nil
}
