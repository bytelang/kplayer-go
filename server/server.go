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
    // s.RegisterValidateRequestFunc(func(r *rpc.RequestInfo, i interface{}) error {
    //     /*
    //     t := reflect.TypeOf(i)
    //     if t.Kind() == reflect.Ptr {
    //         t = t.Elem()
    //     }
    //
    //     newArgs := reflect.New(t)
    //     for i := 0; i < t.NumField(); i++ {
    //         newArgs.FieldByName(t.Field(i).Name).Set(reflect.ValueOf(t.Field(i)))
    //     }
    //
    //      */
    //     validate := validator.New()
    //     return validate.Struct(i)
    // })

    s.RegisterCodec(json.NewCodec(), "application/json")
    if err := s.RegisterService(kprpc.NewResource(mm), ""); err != nil {
        panic(err)
    }
    if err := s.RegisterService(kprpc.NewOutput(mm), ""); err != nil {
        panic(err)
    }
    if err := s.RegisterService(kprpc.NewPlay(mm), ""); err != nil {
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
