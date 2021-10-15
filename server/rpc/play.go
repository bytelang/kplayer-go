package rpc

import (
    "github.com/bytelang/kplayer/module"
    "net/http"

    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/types/core"
    prompt "github.com/bytelang/kplayer/types/core/prompt"
    "github.com/bytelang/kplayer/types/server"
)

// Play rpc
type Play struct {
    mm module.ModuleManager
}

func NewPlay(manager module.ModuleManager) *Play {
    return &Play{mm: manager}
}

// Stop  stop player on idle
func (s *Play) Stop(r *http.Request, args *server.StopPlayArgs, reply *server.StopPlayReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLAYER_STOP, &prompt.EventPromptPlayerStop{}); err != nil {
        return err
    }
    return nil
}
