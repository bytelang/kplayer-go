package rpc

import (
    "fmt"
    "github.com/bytelang/kplayer/core"
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/module/output/provider"
    kpproto "github.com/bytelang/kplayer/types/core"
    "github.com/bytelang/kplayer/types/core/msg"
    prompt "github.com/bytelang/kplayer/types/core/prompt"
    svrproto "github.com/bytelang/kplayer/types/server"
    "github.com/golang/protobuf/proto"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
    "net/http"
    "net/url"
    "os"
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
        Path:   []byte(args.Output.Path),
        Unique: []byte(args.Output.Unique),
    }); err != nil {
        return err
    }

    outputModule := o.mm[provider.ModuleName]
    outputAddMsg := &msg.EventMessageOutputAdd{}

    keeperCtx := module.NewKeeperContext(uuid.New().String(), kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD, func(msg []byte) bool {
        if err := proto.Unmarshal(msg, outputAddMsg); err != nil {
            log.Fatal(err)
        }
        return string(outputAddMsg.Unique) == args.Output.Unique
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

    reply.Output.Path = string(outputAddMsg.Path)
    reply.Output.Unique = string(outputAddMsg.Unique)

    return nil
}
