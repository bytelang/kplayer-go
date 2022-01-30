package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bytelang/kplayer/types/config"
	errortypes "github.com/bytelang/kplayer/types/error"
	"io/ioutil"
	"os"
	"runtime"

	"github.com/bytelang/kplayer/app"
	"github.com/bytelang/kplayer/cmd"
	"github.com/bytelang/kplayer/module"
	"github.com/bytelang/kplayer/server"
	kptypes "github.com/bytelang/kplayer/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	ctx = context.WithValue(ctx, kptypes.ServerCreatorContextKey, server.NewJsonRPCServer())

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

	clientCtx.Viper = v
	if err := v.Unmarshal(clientCtx.Config); err != nil {
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
		d, err := json.Marshal(v.AllSettings()[m.GetModuleName()])
		if err != nil {
			log.Fatal(err)
		}
		modifyData, err := m.InitConfig(clientCtx, d)
		if err != nil {
			log.Fatal(err)
		}
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
	v.SetDefault("play.skip_invalid_resource", true)

	v.SetDefault("play.rpc.on", true)
	v.SetDefault("play.rpc.port", kptypes.DefaultRPCPort)
	v.SetDefault("play.rpc.address", kptypes.DefaultRPCAddress)

	v.SetDefault("play.encode.video_width", 780)
	v.SetDefault("play.encode.video_height", 480)
	v.SetDefault("play.encode.video_fps", 30)
	v.SetDefault("play.encode.audio_channel_layout", 3)
	v.SetDefault("play.encode.audio_channels", 2)
	v.SetDefault("play.encode.audio_sample_rate", 48000)
	v.SetDefault("play.encode.bit_rate", 0)
	v.SetDefault("play.encode.avg_quality", 0)
}
