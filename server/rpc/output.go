package rpc

import (
    "fmt"
    "github.com/bytelang/kplayer/core"
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/types"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    "github.com/bytelang/kplayer/types/core/proto/msg"
    prompt "github.com/bytelang/kplayer/types/core/proto/prompt"
    svrproto "github.com/bytelang/kplayer/types/server"
    log "github.com/sirupsen/logrus"
    "net/http"
    "net/url"
    "os"
)

const outputModuleName = "output"

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
    // validate
    urlParse, err := url.Parse(args.Output.Path)
    if err != nil {
        return err
    }
    allowScheme := map[string]bool{
        "http":  true,
        "https": true,
        "file":  true,
        "rtmp":  true,
    }
    if _, ok := allowScheme[urlParse.Scheme]; !ok {
        return fmt.Errorf("unsupport output resource protocol")
    }
    if urlParse.Scheme == "file" {
        fileInfo, err := os.Stat(args.Output.Path)
        if err != nil {
            return fmt.Errorf("file not exist. path: %s", args.Output.Path)
        }
        if fileInfo.Mode()&(1<<2) != 0 {
            return fmt.Errorf("file don`t have read permission. path: %s", args.Output.Path)
        }
    }

    // send prompt
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_OUTPUT_ADD, &prompt.EventPromptOutputAdd{
        Output: &kpproto.PromptOutput{
            Path:   []byte(args.Output.Path),
            Unique: []byte(args.Output.Unique),
        },
    }); err != nil {
        return err
    }

    outputModule := o.mm[outputModuleName]
    outputAddMsg := &msg.EventMessageOutputAdd{}

    keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD, func(msg []byte) bool {
        types.UnmarshalProtoMessage(msg, outputAddMsg)
        return string(outputAddMsg.Output.Unique) == args.Output.Unique
    })
    defer keeperCtx.Close()

    if err := outputModule.RegisterKeeperChannel(keeperCtx); err != nil {
        return err
    }

    // wait context
    keeperCtx.Wait()

    if outputAddMsg.Error != nil {
        log.Errorf("%s", outputAddMsg.Error)
        return fmt.Errorf("%s", outputAddMsg.Error)
    }

    reply.Output.Path = string(outputAddMsg.Output.Path)
    reply.Output.Unique = string(outputAddMsg.Output.Unique)

    return nil
}
