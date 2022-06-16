package server

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/cmd"
	"github.com/bytelang/kplayer/module"
	outputprovider "github.com/bytelang/kplayer/module/output/provider"
	playprovider "github.com/bytelang/kplayer/module/play/provider"
	pluginprovider "github.com/bytelang/kplayer/module/plugin/provider"
	resourceprovider "github.com/bytelang/kplayer/module/resource/provider"
	"github.com/bytelang/kplayer/types"
	"github.com/go-playground/validator/v10"
	rpcjson "github.com/gorilla/rpc/v2/json"
	"golang.org/x/net/websocket"
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

	s.RegisterCodec(rpcjson.NewCodec(), "application/json")
	if err := s.RegisterService(kprpc.NewResource(mm.GetModule(resourceprovider.ModuleName).(resourceprovider.ProviderI)), ""); err != nil {
		panic(err)
	}
	if err := s.RegisterService(kprpc.NewOutput(mm.GetModule(outputprovider.ModuleName).(outputprovider.ProviderI)), ""); err != nil {
		panic(err)
	}
	if err := s.RegisterService(kprpc.NewPlay(mm.GetModule(playprovider.ModuleName).(playprovider.ProviderI)), ""); err != nil {
		panic(err)
	}
	if err := s.RegisterService(kprpc.NewPlugin(mm.GetModule(pluginprovider.ModuleName).(pluginprovider.ProviderI)), ""); err != nil {
		panic(err)
	}

	// get play module provider
	playProviderInstance := mm.GetModule(playprovider.ModuleName).(playprovider.ProviderI)
	rpcParams := playProviderInstance.GetRPCParams()
	if !rpcParams.On {
		return
	}

	m := http.NewServeMux()
	m.Handle("/rpc", s)
	m.Handle("/websocket", websocket.Handler(wsClientHandler))
	server := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", rpcParams.Address, rpcParams.HttpPort),
		Handler:           m,
		ReadTimeout:       time.Second * 10,
		ReadHeaderTimeout: time.Second * 10,
		WriteTimeout:      time.Second * 10,
	}

	go func() {
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}

		log.Info("rpc server shutdown success")
		stopChan <- true
	}()

	log.WithFields(log.Fields{"address": rpcParams.Address, "port": rpcParams.HttpPort}).Info("rpc server listening")

	<-stopChan
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}

func wsClientHandler(ws *websocket.Conn) {
	sub, err := cmd.SubscribeMessage("server")
	if err != nil {
		log.WithField("error", err).Errorf("subscribe message failed")
		return
	}
	for {
		message := <-sub
		jsonRawMessage, err := types.ParseMessageToJson(message)
		if err != nil {
			log.WithField("error", err).Errorf("message cannot encode to json")
			break
		}

		_, err = ws.Write(jsonRawMessage)
		if err != nil {
			log.WithField("error", err).Errorf("send websocket client failed")
			break
		}
	}
}
