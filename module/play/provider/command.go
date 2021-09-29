package provider

import (
    "github.com/bytelang/kplayer/core"
    "github.com/bytelang/kplayer/module/play/types"
    "github.com/bytelang/kplayer/server"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    "sync"
)

func GetCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   types.ModuleName,
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
            coreKplayer := core.GetLibKplayerInstance()
            if err := coreKplayer.SetOptions("rtmp", 800, 480, 0, 0, 30, 48000, 3, 2); err != nil {
                log.Fatal(err)
            }

            waitGroup := sync.WaitGroup{}
            waitGroup.Add(2)
            serverStopChan := make(chan bool)

            go func() {
                coreKplayer.Run()
                serverStopChan <- true

                waitGroup.Done()
            }()

            go func() {
                server.StartServer(serverStopChan)
                waitGroup.Done()
            }()

            waitGroup.Wait()
            return nil
        },
    }

    return cmd
}
