package provider

import (
	"sync"

	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/server"

	"github.com/bytelang/kplayer/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ModuleName,
		Short: "play category",
		Long:  `App management commands. control kplayer basic status`,
	}

	cmd.AddCommand(StartCommand())

	return cmd
}

func StartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start kplayer",
		Long:  "Start the kplayer application, use '-g' support the daemon mode. on daemon mode, kplayer with creating PID file and same directory only run once.",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get module manager
			var mm module.ModuleManager
			if ptr, err := kptypes.GetCommandContext(cmd, kptypes.ModuleManagerContextKey); err != nil {
				log.Fatal(err)
			} else {
				mm = ptr.(module.ModuleManager)
			}

			// get client ctx
			var clientCtx *kptypes.ClientContext
			if ptr, err := kptypes.GetCommandContext(cmd, kptypes.ClientContextKey); err != nil {
				log.Fatal(err)
			} else {
				clientCtx = ptr.(*kptypes.ClientContext)
			}

			// get server creator
			svrCreator, err := kptypes.GetCommandContext(cmd, kptypes.ServerCreatorContextKey)
			if err != nil {
				log.Fatal(err)
			}

			cfg := clientCtx.Config
			coreKplayer := core.GetLibKplayerInstance()
			if err := coreKplayer.SetOptions(cfg.Play.EncodeModel,
				cfg.Play.Encode.VideoWidth,
				cfg.Play.Encode.VideoHeight,
				cfg.Play.Encode.BitRate,
				cfg.Play.Encode.AvgQuality,
				cfg.Play.Encode.VideoFps,
				cfg.Play.Encode.AudioSampleRate,
				cfg.Play.Encode.AudioChannelLayout,
				cfg.Play.Encode.AudioChannels); err != nil {
				log.Fatal(err)
			}
			coreKplayer.SetCacheOn(cfg.Play.CacheOn)
			coreKplayer.SetSkipInvalidResource(cfg.Play.SkipInvalidResource)

			waitGroup := sync.WaitGroup{}
			waitGroup.Add(2)
			serverStopChan := make(chan bool)

			go func() {
				coreKplayer.Run()
				serverStopChan <- true

				waitGroup.Done()
			}()

			go func() {
				(svrCreator).(server.ServerCreator).StartServer(serverStopChan, mm)
				waitGroup.Done()
			}()

			waitGroup.Wait()
			return nil
		},
	}

	return cmd
}
