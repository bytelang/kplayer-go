package server

import (
	"context"
	"fmt"
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

	// get play module provider
	playProviderInstance := mm[playprovider.ModuleName].(playprovider.ProviderI)
	rpc := playProviderInstance.GetRPCParams()
	if !rpc.On {
		return
	}

	m := http.NewServeMux()
	m.Handle("/rpc", s)
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", rpc.Address, rpc.Port),
		Handler: m,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			panic(err)
		}

		log.Info("rpc server shutdown success")
		stopChan <- true
	}()

	log.WithFields(log.Fields{"address": rpc.Address, "port": rpc.Port}).Info("rpc server listening")

	<-stopChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
