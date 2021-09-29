package rpc

import (
    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/proto"
    prompt "github.com/bytelang/kplayer/proto/prompt"
    "github.com/bytelang/kplayer/server/proto"
    "net/http"
)

type Play struct {
}

func (s *Play) Stop(r *http.Request, args *proto.StopPlayArgs, reply *proto.StopPlayReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EventAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &prompt.EventPromptPlayerStop{
    }); err != nil {
        return err
    }

    reply = &proto.StopPlayReply{}
    return nil
}
