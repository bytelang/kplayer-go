package rpc

import (
    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/proto"
    prompt "github.com/bytelang/kplayer/proto/prompt"
    "github.com/bytelang/kplayer/server/proto"
    "net/http"
)

type Resource struct {
}

func (s *Resource) Add(r *http.Request, args *proto.AddResourceArgs, reply *proto.AddResourceReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EventAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &prompt.EventPromptResourceAdd{
        Path:   args.Path,
        Unique: args.Unique,
    }); err != nil {
        return err
    }

    reply = &proto.AddResourceReply{}
    return nil
}
