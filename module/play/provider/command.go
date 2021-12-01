package provider

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/server"
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
		Short: "play category",
		Long:  `App management commands. control kplayer basic status`,
	}

	cmd.AddCommand(startCommand())
	cmd.AddCommand(stopCommand())
	cmd.AddCommand(statusCommand())

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
				log.WithField("status", "off").Info("kplayer active running on daemon mode")
				return nil
			}
			log.WithFields(log.Fields{"status": "on", "pid": pid}).Info("kplayer active running on daemon mode")
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
				return err
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
		RunE: func(cmd *cobra.Command, args []string) error {
			// daemon mode
			var daemonProc *os.Process
			if cmd.Flag(DaemonMode).Value.String() == DaemonModeYesValue {
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
				for _, m := range mm.Modules {
					m.BeginRunning()
				}
				defer func() {
					for _, m := range mm.Modules {
						m.EndRunning()
					}
				}()

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

	cmd.PersistentFlags().StringP(DaemonMode, "g", DaemonModeDefaultValue, "use daemon mode run kplayer")

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
