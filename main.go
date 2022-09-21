package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/bytelang/kplayer/app"
	"github.com/bytelang/kplayer/cmd"
	"github.com/bytelang/kplayer/module"
	"github.com/bytelang/kplayer/module/play/provider"
	"github.com/bytelang/kplayer/server"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	errortypes "github.com/bytelang/kplayer/types/error"
	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"google.golang.org/protobuf/types/known/anypb"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetReportCaller(true)
	log.SetLevel(log.TraceLevel)
	logFormat := &log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		DisableColors:   false,
		FullTimestamp:   true,
		CallerPrettyfier: func(f *runtime.Frame) (function string, file string) {
			return "", fmt.Sprintf("%s:%d", f.File, f.Line)
		},
	}
	log.SetFormatter(logFormat)
}

func main() {
	rootCmd := cmd.NewRootCmd()

	if err := Execute(rootCmd, app.DefaultConfigFilePath, app.DefaultConfigFileName); err != nil {
		switch e := err.(type) {
		case kptypes.ErrorCode:
			os.Exit(e.Code)
		default:
			os.Exit(1)
		}
	}
}

// Execute execute from flags and commands
func Execute(rootCmd *cobra.Command, defaultHome string, defaultFile string) error {
	rootCmd.PersistentFlags().String(kptypes.FlagLogLevel, log.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
	rootCmd.PersistentFlags().String(kptypes.FlagLogFormat, "plain", "The logging format (json|plain)")
	rootCmd.PersistentFlags().StringP(kptypes.FlagHome, "", defaultHome, "directory for config and data")
	rootCmd.PersistentFlags().StringP(kptypes.FlagConfigFileName, "c", defaultFile, "config file name")
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// set home path
		homePath, err := cmd.Flags().GetString(kptypes.FlagHome)
		if err != nil {
			return err
		}
		if homePath != "" {
			if err := os.Chdir(homePath); err != nil {
				log.WithField("error", err).Fatal("chdir failed")
			}
		}

		// init context
		InitGlobalContextConfig(cmd)
		return nil
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, kptypes.ClientContextKey, kptypes.DefaultClientContext())
	ctx = context.WithValue(ctx, kptypes.ModuleManagerContextKey, app.ModuleManager)
	ctx = context.WithValue(ctx, kptypes.ServerCreatorContextKey, server.NewHttpServer())

	return kptypes.SetCommandContextAndExecute(rootCmd, ctx)
}

func InitGlobalContextConfig(cmd *cobra.Command) {
	mm := cmd.Context().Value(kptypes.ModuleManagerContextKey).(module.ModuleManager)
	clientCtx := cmd.Context().Value(kptypes.ClientContextKey).(*kptypes.ClientContext)

	configFileName, err := kptypes.GetConfigFileName(cmd)
	if err != nil {
		log.Fatal(err)
	}

	// set log level
	logLevel, err := kptypes.GetLogLevel(cmd)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(logLevel)
	if logLevel == log.InfoLevel {
		log.SetReportCaller(false)
	}

	// viper
	v := viper.New()
	v.AddConfigPath(".")
	v.SetConfigType("json")
	if !kptypes.FileExists(configFileName) && !kptypes.FileExists(configFileName+".json") {
		v.SetConfigType("yaml")
	}
	v.SetConfigName(configFileName)

	// skip on init stage
	if cmd.Parent().Use == "init" {
		return
	}

	// load config context in file
	if err := v.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	// set default value
	setDefaultConfig(v)

	// refill resource list proto.any
	// viper decode not support protobuf any unpack
	// prepare constructing a resource list
	var resourceLists []*anypb.Any
	for _, item := range v.Get("resource.lists").([]interface{}) {
		switch itemResource := item.(type) {
		case string:
			singleResource := &config.SingleResource{
				Path: itemResource,
			}
			any, err := ptypes.MarshalAny(singleResource)
			if err != nil {
				log.WithField("error", err).Fatal("unmarshal any failed")
			}
			resourceLists = append(resourceLists, any)
		case map[string]interface{}:
			var found bool = false

			bytes, err := json.Marshal(item)
			if err != nil {
				log.WithField("error", err).Fatal("unmarshal resource item failed")
			}
			// single resource
			{
				singleResource := &config.SingleResource{}
				if err := kptypes.UnmarshalProtoMessageContinue(string(bytes), singleResource); err == nil {
					any, err := ptypes.MarshalAny(singleResource)
					if err != nil {
						log.WithField("error", err).Fatal("unmarshal any failed")
					}
					resourceLists = append(resourceLists, any)
					found = true
				}
			}

			// mix resource
			{
				mixResource := &config.MixResource{}
				if err := kptypes.UnmarshalProtoMessageContinue(string(bytes), mixResource); err == nil {
					any, err := ptypes.MarshalAny(mixResource)
					if err != nil {
						log.WithField("error", err).Fatal("unmarshal any failed")
					}
					resourceLists = append(resourceLists, any)
					found = true
				}
			}

			if !found {
				log.WithField("content", item).Warn("unrecognized resource structure")
			}
		}
	}
	v.Set("resource.lists", map[string]interface{}{})

	// unmarshal config
	clientCtx.Viper = v
	if err := v.Unmarshal(clientCtx.Config); err != nil {
		log.Fatal(err)
	}

	// set resource list
	clientCtx.Config.Resource.Lists = resourceLists

	// custom config
	{
		// add generate cache config
		if cmd.Flag(provider.FlagGenerateCache) != nil {
			if cmd.Flag(provider.FlagGenerateCache).Value.String() == provider.FlagYesValue {
				clientCtx.Config.Play.PlayModel = strings.ToLower(config.PLAY_MODEL_name[int32(config.PLAY_MODEL_LIST)])
				clientCtx.Config.Play.EncodeModel = strings.ToLower(config.ENCODE_MODEL_name[int32(config.ENCODE_MODEL_FILE)])
				clientCtx.Config.Output.Lists = nil
				clientCtx.Config.Play.CacheOn = true
			}
		}
	}

	reInitConfig, err := json.Marshal(clientCtx.Config)
	if err != nil {
		log.Fatal(err)
	}
	if err := v.ReadConfig(bytes.NewBuffer(reInitConfig)); err != nil {
		log.Fatal(err)
	}

	// validate global config
	if err := ValidateConfig(clientCtx.Config); err != nil {
		log.Fatal(err)
	}

	// init module
	for _, item := range mm.OrderInitConfig {
		m := mm.GetModule(item)

		// init config and set default value
		d, err := json.Marshal(clientCtx.Config)
		if err != nil {
			log.Fatal(err)
		}
		modifyData, err := m.InitConfig(clientCtx, []byte(gjson.Parse(string(d)).Get(m.GetModuleName()).String()))
		if err != nil {
			log.Fatal(err)
		}

		// validator
		validate := validator.New()
		if err := validate.Struct(modifyData); err != nil {
			log.Fatal(err)
		}

		// set modify data
		v.Set(m.GetModuleName(), modifyData)

		// validate config
		if err := m.ValidateConfig(); err != nil {
			log.Fatal(err)
		}
	}

	// set context before module modify
	if err := v.Unmarshal(clientCtx.Config); err != nil {
		log.Fatal(err)
	}
}

func ValidateConfig(config *config.KPConfig) error {
	if config.Version != app.ConfigVersion {
		return errortypes.VersionInvalidMainError
	}

	// load user token file
	if config.TokenPath != "" {
		if !kptypes.FileExists(config.TokenPath) {
			return errortypes.TokenFileNotFoundMainError
		}

		fileContent, err := ioutil.ReadFile(config.TokenPath)
		if err != nil {
			return err
		}

		if err := kptypes.LoadClientToken(string(fileContent)); err != nil {
			log.Fatal(err)
		}
	}

	return nil
}

func setDefaultConfig(v *viper.Viper) {
	v.SetDefault("play.start_point", 1)
	v.SetDefault("play.play_model", "list")
	v.SetDefault("play.encode_model", "rtmp")
	v.SetDefault("play.cache_on", false)
	v.SetDefault("play.cache_uncheck", false)
	v.SetDefault("play.skip_invalid_resource", false)
	v.SetDefault("play.delay_queue_size", 50)
	v.SetDefault("play.fill_strategy", "tile")

	v.SetDefault("play.rpc.on", true)
	v.SetDefault("play.rpc.http_port", kptypes.DefaultHttpPort)
	v.SetDefault("play.rpc.grpc_port", kptypes.DefaultRPCPort)
	v.SetDefault("play.rpc.address", kptypes.DefaultRPCAddress)

	v.SetDefault("play.encode.video_width", 854)
	v.SetDefault("play.encode.video_height", 480)
	v.SetDefault("play.encode.video_fps", 25)
	v.SetDefault("play.encode.audio_channel_layout", 3)
	v.SetDefault("play.encode.audio_channels", 2)
	v.SetDefault("play.encode.audio_sample_rate", 44100)
	v.SetDefault("play.encode.bit_rate", 0)
	v.SetDefault("play.encode.avg_quality", 0)

	// auth
	v.SetDefault("auth.auth_on", false)
}
