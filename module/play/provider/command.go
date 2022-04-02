package provider

import (
	"fmt"
	"github.com/bytelang/kplayer/types/client"
	"github.com/bytelang/kplayer/types/config"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	kpserver "github.com/bytelang/kplayer/types/server"
	"github.com/sevlyar/go-daemon"

	"github.com/bytelang/kplayer/core"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const (
	pidFilePath = "log/kplayer.pid"
	logFilePath = "log/kplayer.log"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ModuleName,
		Short: "Play category",
		Long:  `App management commands. control kplayer basic status`,
	}

	cmd.AddCommand(startCommand())
	cmd.AddCommand(stopCommand())
	cmd.AddCommand(statusCommand())
	cmd.AddCommand(durationCommand())
	cmd.AddCommand(pauseCommand())
	cmd.AddCommand(continueCommand())
	cmd.AddCommand(skipCommand())
	cmd.AddCommand(versionCommand())

	return cmd
}

func statusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Print kplayer status",
		Long:  "Get the kplayer application running status",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := getPID()
			if err != nil {
				log.WithField("status", "off").Info("kplayer not running on daemon mode")
				return nil
			}
			log.WithFields(log.Fields{"status": "on", "pid": pid}).Info("kplayer active running on daemon mode")
			return nil
		},
	}
	return cmd
}

func durationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "duration",
		Short: "get player duration status",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.PlayDurationReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Play.Duration", &kpserver.PlayDurationArgs{}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	return cmd
}

func pauseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause",
		Short: "pause player",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.PlayPauseReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Play.Pause", &kpserver.PlayPauseArgs{}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	return cmd
}

func continueCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "continue",
		Short: "continue player",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.PlayPauseReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Play.Pause", &kpserver.PlayPauseArgs{}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	return cmd
}

func skipCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skip",
		Short: "skip play current resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.PlaySkipReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Play.Skip", &kpserver.PlaySkipArgs{}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	return cmd
}

func versionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "get Information play",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.PlayInformationReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Play.Information", &kpserver.PlayInformationArgs{}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	return cmd
}

func stopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Start kplayer",
		Long:  "Stop the kplayer application. only effective in daemon mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			pid, err := getPID()
			if err != nil {
				log.WithField("error", err).Error("stop failed")
				return nil
			}

			// kill process
			if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
				log.WithField("error", err).Error("kill process failed")
				return err
			}

			log.Info("kplayer stop success")
			return nil
		},
	}

	return cmd
}

func startCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start kplayer",
		Long:  "Start the kplayer application, use '-g' support the daemon mode. on daemon mode, kplayer with creating PID file and same directory only run once.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return kptypes.MkDir("log")
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// daemon mode
			var daemonProc *os.Process
			if cmd.Flag(FlagDaemonMode).Value.String() == FlagYesValue {
				var err error

				if pid, err := getPID(); err == nil {
					log.WithField("pid", pid).Error("kplayer start failed. kplayer is running")
					return nil
				}

				cntxt := &daemon.Context{
					PidFileName: pidFilePath,
					PidFilePerm: 0644,
					LogFileName: logFilePath,
					LogFilePerm: 0644,
					WorkDir:     "./",
					Env:         os.Environ(),
					Args:        cmd.Flags().Args(),
					Umask:       027,
				}
				daemonProc, err = cntxt.Reborn()
				if err != nil {
					log.WithField("error", err).Fatal("execute daemon mode failed")
				}
				if daemonProc != nil {
					log.Info("kplayer start success on daemon mode")
					return nil
				}
			} else {
				// not daemon mode
				// write pid to file
				_ = os.Mkdir(filepath.Dir(pidFilePath), os.ModePerm)
				f, err := os.OpenFile(pidFilePath, os.O_CREATE|os.O_RDWR, 0666)
				if err != nil {
					log.WithField("error", err).Fatal("open pid file failed")
					return nil
				}
				if _, err := io.WriteString(f, strconv.Itoa(os.Getpid())); err != nil {
					log.WithField("error", err).Fatal("write pid file failed")
				}
			}
			defer func() {
				if daemonProc != nil {
					_ = daemonProc.Release()
				}
			}()

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
			if cmd.Flag(FlagGenerateCache).Value.String() == FlagYesValue {
				cfg.Play.EncodeModel = config.ENCODE_MODEL_name[int32(config.ENCODE_MODEL_FILE)]
				cfg.Play.CacheOn = true
				log.Info("running on generate cache model")
			}

			coreKplayer := core.GetLibKplayerInstance()
			if err := coreKplayer.SetOptions(cfg.Play.EncodeModel,
				cfg.Play.Encode.VideoWidth,
				cfg.Play.Encode.VideoHeight,
				cfg.Play.Encode.BitRate,
				cfg.Play.Encode.AvgQuality,
				cfg.Play.Encode.VideoFps,
				cfg.Play.Encode.AudioSampleRate,
				cfg.Play.Encode.AudioChannelLayout,
				cfg.Play.Encode.AudioChannels,
				cfg.Play.DelayQueueSize); err != nil {
				log.Fatal(err)
			}
			coreKplayer.SetCacheOn(cfg.Play.CacheOn)
			coreKplayer.SetSkipInvalidResource(cfg.Play.SkipInvalidResource)

			serverStopChan := make(chan bool)

			var coreLogLevel int = 0
			level, err := cmd.Flags().GetString(kptypes.FlagLogLevel)
			if err != nil {
				log.Fatal(err)
			}
			logLevel, err := log.ParseLevel(level)
			switch logLevel {
			case log.TraceLevel:
				coreLogLevel = 2
			case log.DebugLevel:
				coreLogLevel = 1
			default:
				coreLogLevel = 0
			}

			// module option
			moduleOptions := []module.ModuleOption{}
			if cmd.Flag(FlagGenerateCache).Value.String() == FlagYesValue {
				moduleOptions = append(moduleOptions, module.ModuleOptionGenerateCache)
			}

			// knock api
			timeTicker := time.NewTicker(time.Minute * KnockIntervalMinutes)
			defer timeTicker.Stop()
			go func() {
				maxRetriesCount := KnockMaxRetries
				currentRetriesCount := 0
				for {
					if currentRetriesCount > maxRetriesCount {
						log.Fatal("knock failed. cannot connection api server on max retries1")
					}

					<-timeTicker.C
					if err := kptypes.Knock(); err != nil {
						currentRetriesCount = currentRetriesCount + 1
						continue
					}

					log.Debug("knock success")
					currentRetriesCount = 0
				}
			}()

			go func() {
				(svrCreator).(kpserver.ServerCreator).StartServer(serverStopChan, mm)
			}()

			// start core
			{
				for _, m := range mm.Modules {
					m.BeginRunning(moduleOptions...)
				}
				defer func() {
					for _, m := range mm.Modules {
						m.EndRunning()
					}
				}()

				coreKplayer.SetLogLevel("log/core.log", coreLogLevel)
				coreKplayer.Run()
				serverStopChan <- true
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolP(FlagDaemonMode, "d", false, "use daemon mode run kplayer")
	cmd.PersistentFlags().BoolP(FlagGenerateCache, "g", false, "only generate file cache. not push to output")

	return cmd
}

func getPID() (int, error) {
	pidFile, err := os.Open(pidFilePath)
	if err != nil {
		return 0, fmt.Errorf("can not found pid file. error: %s", err)
	}
	defer pidFile.Close()

	data, err := ioutil.ReadAll(pidFile)
	if err != nil {
		return 0, fmt.Errorf("read pid file failed. error: %s", err)
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, fmt.Errorf("pid invalid. error: %s", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, fmt.Errorf("kplayer not running on daemon mode")
	}
	if err := process.Signal(syscall.Signal(0)); err != nil {
		return 0, fmt.Errorf("kplayer not running on daemon mode")
	}

	return pid, nil
}
