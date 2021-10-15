package cmd

import (
    "github.com/bytelang/kplayer/app"
    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/types/core"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

const (
    DefaultConfigFileName = "kplayer.yaml"
    DefaultConfigFilePath = "./"
)

func NewRootCmd() *cobra.Command {

    // init kplayer
    coreKplayer := core.GetLibKplayerInstance()
    coreKplayer.SetCallBackMessage(messageConsumer)

    // init command
    rootCmd := &cobra.Command{
        Use:   "kplayer",
        Short: "kplayer launch application1",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            cmd.SetOut(cmd.OutOrStdout())
            cmd.SetErr(cmd.ErrOrStderr())
        },
    }

    initRootCmd(rootCmd)

    return rootCmd
}

func initRootCmd(rootCmd *cobra.Command) {
    // add module command
    app.AddCommands(rootCmd)
}

func messageConsumer(message *kpproto.KPMessage) {
    log.Debug("receive broadcast message: ", message.Action)

    for _, item := range app.ModuleManager {
        item.Trigger(message)
    }
}
