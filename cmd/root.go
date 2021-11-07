package cmd

import (
    "fmt"
    "github.com/bytelang/kplayer/app"
    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    terminal "github.com/wayneashleyberry/terminal-dimensions"
)

var (
    MAJOR_TAG  string = "<MAJOR_TAG>"
    MAJOR_HASH string = "<MAJOR_HASH>"
    WebSite    string = "<WEB_SITE>"
)

const terminalCharsetMaxCount uint = 115

func NewRootCmd() *cobra.Command {
    // init core
    coreKplayer := core.GetLibKplayerInstance()
    coreKplayer.SetCallBackMessage(messageConsumer)

    // get core information
    info := coreKplayer.GetInformation()

    terminalWidth, _, err := terminal.Dimensions()
    if err != nil {
        log.Fatal(err)
    }
    if terminalWidth > terminalCharsetMaxCount {
        terminalWidth = terminalCharsetMaxCount
    }

    // init command
    shortDesc := fmt.Sprintf(`kplayer for golang %s(%s) Copyright (c) %s the ByteLang Studio (%s)
  libkplayer version: %s plugin version: %s license version: %s 
  toolchains %s C++ Standard %s on %s
  build with build-chains %s, type with %s
  Hope you have a good experience.
`,
        MAJOR_TAG,
        MAJOR_HASH,
        info.Copyright,
        WebSite,
        info.MajorVersion,
        info.PluginVersion,
        info.LicenseVersion,
        info.BuildChains,
        info.BuildType,
        info.ToolChains,
        info.CppStd,
        info.ArchiveVersion,
    )

    for i := 0; uint(i) < terminalWidth; i++ {
        shortDesc = shortDesc + "-"
    }
    shortDesc = shortDesc + "\n"

    rootCmd := &cobra.Command{
        Use:   app.AppName,
        Short: shortDesc,
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
