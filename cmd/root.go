package cmd

import (
    "github.com/bytelang/kplayer/app"
    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/proto"
    "github.com/bytelang/kplayer/proto/prompt"
    "github.com/bytelang/kplayer/types"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
)

const (
    DefaultConfigFileName = "kplayer.yaml"
    DefaultConfigFilePath = "./"
)

var clientCtx *types.ClientContext
var kplayerApp *app.KplayerApp

func NewRootCmd() *cobra.Command {

    // init kplayer
    coreKplayer := core.GetLibKplayerInstance()
    coreKplayer.SetCallBackMessage(messageConsumer)

    // init command
    rootCmd := &cobra.Command{
        Use:   "kplayer",
        Short: "kplayer launch application",
        PreRun: func(cmd *cobra.Command, args []string) {
            clientCtx, err := types.GetCommandContext(cmd, types.ClientContextKey)
            if err != nil {
                panic(err)
            }
            kplayerApp, err := types.GetCommandContext(cmd, types.AppContextKey)
            if err != nil {
                panic(err)
            }

            // assignment global context
            clientCtx = clientCtx.(*types.ClientContext)
            kplayerApp = kplayerApp.(*app.KplayerApp)
        },
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
    app.ModuleManager.AddCommands(rootCmd)
}

func messageConsumer(message *kpproto.KPMessage) {
    log.Debug("receive broadcast message: ", message.Action)

    // global core
    coreKplayer := core.GetLibKplayerInstance()

    for _, item := range app.ModuleManager.Manager {
        log.Info(item.GetModuleName())
    }

    var err error
    switch message.Action {
    case kpproto.EventAction_EVENT_MESSAGE_ACTION_PLAYER_STARTED:
        // add output
        if err = coreKplayer.SendPrompt(kpproto.EventAction_EVENT_PROMPT_ACTION_OUTPUT_ADD, &prompt.EventPromptOutputAdd{
            Path:   "output.flv",
            Unique: "test",
        }); err != nil {
            break
        }

        // add input
        if err = coreKplayer.SendPrompt(kpproto.EventAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &prompt.EventPromptResourceAdd{
            Path:   "/Users/kangkai/smart/video/short.flv",
            Unique: "qflasd",
        }); err != nil {
            break
        }
    case kpproto.EventAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY:
        err = coreKplayer.SendPrompt(kpproto.EventAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &prompt.EventPromptPlayerStop{})
    }

    if err != nil {
        log.Errorf("send prompt command failed. error: %s", err)
    }
}
