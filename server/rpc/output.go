package rpc

import (
    "net/http"

    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/proto"
    prompt "github.com/bytelang/kplayer/proto/prompt"
    "github.com/bytelang/kplayer/server/proto"
)

// Output rpc
type Output struct {
}

// Add add output to core player
func (o *Output) Add(r *http.Request, args *proto.AddOutputArgs, reply *proto.AddOutputReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EventAction_EVENT_PROMPT_ACTION_OUTPUT_ADD, &prompt.EventPromptOutputAdd{
        Path:   args.Path,
        Unique: args.Unique,
    }); err != nil {
        return err
    }



    return nil
}
