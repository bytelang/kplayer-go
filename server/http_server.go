package server

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/module"
	outputprovider "github.com/bytelang/kplayer/module/output/provider"
	playprovider "github.com/bytelang/kplayer/module/play/provider"
	pluginprovider "github.com/bytelang/kplayer/module/plugin/provider"
	resourceprovider "github.com/bytelang/kplayer/module/resource/provider"
	kptypes "github.com/bytelang/kplayer/types"
	autherror "github.com/bytelang/kplayer/types/error"
	"github.com/bytelang/kplayer/types/server"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
)

const AUTHORIZATION_METADATA_KEY = "Authorization"

type httpServer struct {
	authOn    bool
	authToken string
}

func NewHttpServer() *httpServer {
	return &httpServer{}
}

var _ server.ServerCreator = &httpServer{}

func (h *httpServer) StartServer(stopChan chan bool, mm module.ModuleManager, authOn bool, authToken string) {
	h.authToken = authToken
	h.authOn = authOn

	// modules
	playModule := mm.GetModule(playprovider.ModuleName).(playprovider.ProviderI)
	outputModule := mm.GetModule(outputprovider.ModuleName).(outputprovider.ProviderI)
	pluginModule := mm.GetModule(pluginprovider.ModuleName).(pluginprovider.ProviderI)
	resourceModule := mm.GetModule(resourceprovider.ModuleName).(resourceprovider.ProviderI)

	grpcEndpoint := fmt.Sprintf("%s:%d", playModule.GetRPCParams().Address, playModule.GetRPCParams().GrpcPort)
	httpEndpoint := fmt.Sprintf("%s:%d", playModule.GetRPCParams().Address, playModule.GetRPCParams().HttpPort)

	opts := []grpc_recovery.Option{
		grpc_recovery.WithRecoveryHandler(func(p interface{}) (err error) {
			return status.Errorf(codes.Unknown, "panic triggered: %v", p)
		}),
	}
	reqValidatorInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		if err := kptypes.ValidateStructor(req); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}

		if len(h.authToken) != 0 {
			md, exist := metadata.FromIncomingContext(ctx)
			if !exist {
				return nil, autherror.AuthTokenNotExists
			}
			token := md.Get(AUTHORIZATION_METADATA_KEY)
			if h.authOn && token[0] != h.authToken {
				return nil, status.Error(codes.Unauthenticated, autherror.AuthTokenInvalid.Error())
			}
		}
		return handler(ctx, req)
	}

	grpcSvc := grpc.NewServer(
		grpc_middleware.WithUnaryServerChain(
			grpc_recovery.UnaryServerInterceptor(opts...),
			grpc_middleware.ChainUnaryServer(reqValidatorInterceptor),
		),
		grpc_middleware.WithStreamServerChain(grpc_recovery.StreamServerInterceptor(opts...)),
	)
	httpSvc := http.Server{}

	go func() {
		// grpc server
		listen, err := net.Listen("tcp", grpcEndpoint)
		if err != nil {
			log.WithField("error", err).Fatal("start grpc gateway server failed")
		}

		server.RegisterOutputGreeterServer(grpcSvc, outputModule)
		server.RegisterPlayGreeterServer(grpcSvc, playModule)
		server.RegisterPluginGreeterServer(grpcSvc, pluginModule)
		server.RegisterResourceGreeterServer(grpcSvc, resourceModule)

		err = grpcSvc.Serve(listen)
		if err != nil {
			log.WithField("error", err).Fatal("start grpc gateway server failed")
		}
		log.Info("rpc server shutdown success")
	}()

	go func() {
		// grpc-gateway server
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// http server
		listen, err := net.Listen("tcp", httpEndpoint)
		if err != nil {
			log.WithField("error", err).Fatal("start grpc gateway server failed")
		}

		// Register gRPC server endpoint
		mux := runtime.NewServeMux(
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{OrigName: true, EmitDefaults: true}),
			runtime.WithProtoErrorHandler(protoErrorHandle),
			runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
				switch key {
				case AUTHORIZATION_METADATA_KEY:
					return key, true
				}
				return "", false
			}),
		)
		opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
		err = server.RegisterPlayGreeterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			log.WithField("error", err).Panic("register grpc gateway server failed")
		}
		err = server.RegisterOutputGreeterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			log.WithField("error", err).Panic("register grpc gateway server failed")
		}
		err = server.RegisterPluginGreeterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			log.WithField("error", err).Panic("register grpc gateway server failed")
		}
		err = server.RegisterResourceGreeterHandlerFromEndpoint(ctx, mux, grpcEndpoint, opts)
		if err != nil {
			log.WithField("error", err).Panic("register grpc gateway server failed")
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
	_ = httpSvc.Close()
}

func protoErrorHandle(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
	s, ok := status.FromError(err)
	if !ok {
		s = status.New(codes.Unknown, err.Error())
	}

	// set header
	writer.Header().Del("Trailer")
	writer.Header().Set("Context-Type", marshaler.ContentType())

	// set content
	body := &struct {
		InternalCode codes.Code
		Message      string
		Details      []interface{}
	}{}

	body.InternalCode = s.Code()
	body.Message = err.Error()
	body.Details = s.Details()

	buf, merr := marshaler.Marshal(body)
	if merr != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		_, _ = writer.Write([]byte(merr.Error()))
		return
	}

	// set status
	writer.WriteHeader(runtime.HTTPStatusFromCode(s.Code()))
	_, _ = writer.Write(buf)
}
