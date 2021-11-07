package cmd

import (
    "fmt"
    "github.com/bytelang/kplayer/app"
    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

var (
    MAJOR_TAG string = "<MAJOR_TAG>"
    WebSite   string = "<WEB_SITE>"
)

func NewRootCmd() *cobra.Command {
    // init core
    coreKplayer := core.GetLibKplayerInstance()
    coreKplayer.SetCallBackMessage(messageConsumer)

    // get core information
    info := coreKplayer.GetInformation()

    // init command
    rootCmd := &cobra.Command{
        Use: app.AppName,
        Short: fmt.Sprintf(`kplayer for golang major version %s Copyright (c) %s the ByteLang Studio (%s)
  core libkplayer version: %s, plugin version: %s, license version: %s 
  build with buildchains %s, toolchains %s, type with %s on %s
  Hope you have a good experience.
`,
            MAJOR_TAG,
            info.Copyright,
            WebSite,
            info.MajorVersion,
            info.PluginVersion,
            info.LicenseVersion,
            info.BuildChains,
            info.ToolChains,
            info.BuildType,
            info.ArchiveVersion,
        ),
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
    log.WithFields(log.Fields{"action": kpproto.EventAction_name[int32(message.Action)]}).Debug("receive broadcast message")

    var copyMsg kpproto.KPMessage
    for _, item := range app.ModuleManager {
        copyMsg = *message
        item.ParseMessage(&copyMsg)

        copyMsg = *message
        item.TriggerMessage(&copyMsg)
    }
}
