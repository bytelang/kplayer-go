package provider

import (
    "github.com/bytelang/kplayer/client"
    "github.com/bytelang/kplayer/core"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

var clientCtx *client.ClientContext

func GetCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "app",
        Short: "kplayer application command",
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
        PreRun: func(cmd *cobra.Command, args []string) {
            ctx, err := client.GetClientContext(cmd)
            if err != nil {
                panic(err)
            }
            clientCtx = ctx
        },
        RunE: func(cmd *cobra.Command, args []string) error {
            coreKplayer := core.GetLibKplayerInstance()
            if err := coreKplayer.SetOptions("rtmp", 800, 480, 0, 0, 30, 48000, 3, 2); err != nil {
                log.Fatal(err)
            }
            coreKplayer.Run()
            return nil
        },
    }

    return cmd
}
