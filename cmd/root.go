package cmd

import (
    "github.com/bytelang/kplayer/app"
    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
    // init core
    coreKplayer := core.GetLibKplayerInstance()
    coreKplayer.SetCallBackMessage(messageConsumer)

    // init command
    rootCmd := &cobra.Command{
        Use:   app.AppName,
        Short: "launch application",
        PersistentPreRun: func(cmd *cobra.Command, args []string) {
            cmd.SetOut(cmd.OutOrStdout())
            cmd.SetErr(cmd.ErrOrStderr())
        },
    }

    initRootCmd(rootCmd)

    return rootCmd
}

func initRootCmd(rootCmd *cobra.Command) {
    // add init command
    rootCmd.AddCommand(app.AddInitCommands())

    // add module command
    app.AddCommands(rootCmd)
}

func messageConsumer(message *kpproto.KPMessage) {
    log.Debug("receive broadcast message: ", message.Action)

    var copyMsg kpproto.KPMessage
    for _, item := range app.ModuleManager {
        copyMsg = *message
        item.ParseMessage(&copyMsg)

        copyMsg = *message
        item.TriggerMessage(&copyMsg)
    }
}
