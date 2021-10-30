package rpc

import (
    "fmt"
    "github.com/bytelang/kplayer/module/output/provider"
    svrproto "github.com/bytelang/kplayer/types/server"
    "net/http"
    "net/url"
    "os"
)

// Output rpc
type Output struct {
    pi provider.ProviderI
}

func NewOutput(pi provider.ProviderI) *Output {
    return &Output{pi: pi}
}

// Add add output to core player
func (o *Output) Add(r *http.Request, args *svrproto.OutputAddArgs, reply *svrproto.OutputAddReply) error {
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

    // call provider add
    addResource, err := o.pi.OutputAdd(args)
    if err != nil {
        return err
    }

    reply.Output = addResource.Output

    return nil
}

// Remove
func (o *Output) Remove(r *http.Request, args *svrproto.OutputRemoveArgs, reply *svrproto.OutputRemoveReply) error {
    removeResource, err := o.pi.OutputRemove(args)
    if err != nil {
        return err
    }

    reply.Output = removeResource.Output
    return nil
}

// List
func (o *Output) List(r *http.Request, args *svrproto.OutputListArgs, reply *svrproto.OutputListReply) error {
    listResource, err := o.pi.OutputList(args)
    if err != nil {
        return err
    }

    reply.Outputs = listResource.Outputs
    return nil
}