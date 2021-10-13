package server

import (
    "context"
    "github.com/bytelang/kplayer/module"
    "net/http"
    "time"

    kprpc "github.com/bytelang/kplayer/server/rpc"
    "github.com/gorilla/rpc/v2"
    "github.com/gorilla/rpc/v2/json"
    log "github.com/sirupsen/logrus"
)

const (
    address = "0.0.0.0:4156"
)

// StartServer start rpc server
func StartServer(stopChan chan bool, mm module.ModuleManager) {
    s := rpc.NewServer()
    s.RegisterCodec(json.NewCodec(), "application/json")
    if err := s.RegisterService(kprpc.NewResource(mm), ""); err != nil {
        panic(err)
    }
    if err := s.RegisterService(&kprpc.Play{}, ""); err != nil {
        panic(err)
    }
    if err := s.RegisterService(&kprpc.Output{}, ""); err != nil {
        panic(err)
    }

    m := http.NewServeMux()
    m.Handle("/rpc", s)
    server := &http.Server{
        Addr:    address,
        Handler: m,
    }

    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            panic(err)
        }

        log.Info("RPC server shutdown success.")
        stopChan <- true
    }()

    log.Info("RPC server listening on: ", address)

    <-stopChan
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := server.Shutdown(ctx); err != nil {
        panic(err)
    }
}
