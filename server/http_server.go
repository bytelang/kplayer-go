package server

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/module"
	playprovider "github.com/bytelang/kplayer/module/play/provider"
	"github.com/bytelang/kplayer/types/server"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"net/http"
)

type httpServer struct {
}

func NewHttpServer() *httpServer {
	return &httpServer{}
}

var _ server.ServerCreator = &httpServer{}

func (h httpServer) StartServer(stopChan chan bool, mm module.ModuleManager) {
	playModule := mm.GetModule(playprovider.ModuleName).(playprovider.ProviderI)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	grpcEndpoint := fmt.Sprintf("%s:%d", playModule.GetRPCParams().Address, playModule.GetRPCParams().GrpcPort)
	httpEndpoint := fmt.Sprintf("%s:%d", playModule.GetRPCParams().Address, playModule.GetRPCParams().HttpPort)
	grpcSvc := grpc.NewServer()
	httpSvc := http.Server{}

	go func() {
		// grpc server
		listen, err := net.Listen("tcp", grpcEndpoint)
		if err != nil {
			log.WithField("error", err).Panic("start grpc gateway server failed")
		}

		server.RegisterPlayGreeterServer(grpcSvc, playModule)

		err = grpcSvc.Serve(listen)
		if err != nil {
			log.WithField("error", err).Panic("start grpc gateway server failed")
		}
		log.Info("rpc server shutdown success")
	}()

	go func() {
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// http server
		listen, err := net.Listen("tcp", httpEndpoint)
		if err != nil {
			log.WithField("error", err).Panic("start grpc gateway server failed")
		}

		// Register gRPC server endpoint
		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		err = server.RegisterPlayGreeterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			log.WithField("error", err).Panic("start grpc gateway server failed")
		}

		httpSvc.Handler = mux
		httpSvc.Addr = httpEndpoint

		// Start http server
		if err := httpSvc.Serve(listen); err != nil {
			log.WithField("error", err).Panic("start grpc gateway server failed")
		}
	}()

	log.WithFields(log.Fields{"address": playModule.GetRPCParams().Address, "port": playModule.GetRPCParams().GrpcPort}).Info("grpc server listening")
	log.WithFields(log.Fields{"address": playModule.GetRPCParams().Address, "port": playModule.GetRPCParams().HttpPort}).Info("http server listening")

	<-stopChan
	grpcSvc.Stop()
	httpSvc.Close()
}
