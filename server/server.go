package server

import (
	"context"
	"github.com/bytelang/kplayer/module"
	outputprovider "github.com/bytelang/kplayer/module/output/provider"
	playprovider "github.com/bytelang/kplayer/module/play/provider"
	pluginprovider "github.com/bytelang/kplayer/module/plugin/provider"
	resourceprovider "github.com/bytelang/kplayer/module/resource/provider"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/rpc/v2/json"
	"net/http"
	"time"

	kprpc "github.com/bytelang/kplayer/server/rpc"
	"github.com/gorilla/rpc/v2"
	log "github.com/sirupsen/logrus"
)

const (
	address = "0.0.0.0:4156"
)

type jsonRPCServer struct {
}

func NewJsonRPCServer() *jsonRPCServer {
	return &jsonRPCServer{}
}

// StartServer start rpc server
func (jrs *jsonRPCServer) StartServer(stopChan chan bool, mm module.ModuleManager) {
	s := rpc.NewServer()
	s.RegisterValidateRequestFunc(func(r *rpc.RequestInfo, i interface{}) error {
		validate := validator.New()
		return validate.Struct(i)
	})

	s.RegisterCodec(json.NewCodec(), "application/json")
	if err := s.RegisterService(kprpc.NewResource(mm[resourceprovider.ModuleName].(resourceprovider.ProviderI)), ""); err != nil {
		panic(err)
	}
	if err := s.RegisterService(kprpc.NewOutput(mm[outputprovider.ModuleName].(outputprovider.ProviderI)), ""); err != nil {
		panic(err)
	}
	if err := s.RegisterService(kprpc.NewPlay(mm[playprovider.ModuleName].(playprovider.ProviderI)), ""); err != nil {
		panic(err)
	}
	if err := s.RegisterService(kprpc.NewPlugin(mm[pluginprovider.ModuleName].(pluginprovider.ProviderI)), ""); err != nil {
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

	log.Infof("RPC server listening on: %s", address)

	<-stopChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
