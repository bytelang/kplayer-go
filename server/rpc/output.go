package rpc

import (
    "fmt"
    "github.com/bytelang/kplayer/module"
    outputype "github.com/bytelang/kplayer/module/output/types"
    "github.com/bytelang/kplayer/proto/msg"
    "github.com/google/uuid"
    "net/http"

    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/proto"
    prompt "github.com/bytelang/kplayer/proto/prompt"
    svrproto "github.com/bytelang/kplayer/server/proto"
)

// Output rpc
type Output struct {
    mm module.ModuleManager
}

func NewOutput(manager module.ModuleManager) *Output {
    return &Output{mm: manager}
}

// Add add output to core player
func (o *Output) Add(r *http.Request, args *svrproto.AddOutputArgs, reply *svrproto.AddOutputReply) error {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_OUTPUT_ADD, &prompt.EventPromptOutputAdd{
        Path:   args.Output.Path,
        Unique: args.Output.Unique,
    }); err != nil {
        return err
    }

    outputModule := o.mm[outputype.ModuleName]
    keeperCtx := module.NewKeeperContext(uuid.New().String(), kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD)
    defer keeperCtx.Close()

    if err := outputModule.RegisterKeeperChannel(keeperCtx); err != nil {
        return err
    }

    // wait context
    outputAddMsg := &msg.EventMessageOutputAdd{}
    if err := keeperCtx.Wait(outputAddMsg); err != nil {
        return fmt.Errorf("messge type invalid")
    }

    reply.Output.Path = outputAddMsg.Path
    reply.Output.Unique = outputAddMsg.Unique

    return nil
}
