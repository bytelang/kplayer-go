package cmd

import (
	"fmt"
	"sync"

	"github.com/bytelang/kplayer/app"
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/types"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	terminal "github.com/wayneashleyberry/terminal-dimensions"
)

const terminalCharsetMaxCount uint = 115

var subscribeCollector map[string]chan kpproto.KPMessage
var subscribeMutex sync.Mutex

func init() {
	subscribeCollector = make(map[string]chan kpproto.KPMessage)
}

func NewRootCmd() *cobra.Command {
	// init core
	coreKplayer := core.GetLibKplayerInstance()
	coreKplayer.SetCallBackMessage(messageConsumer)

	// get core information
	info := coreKplayer.GetInformation()

	terminalWidth, _, err := terminal.Dimensions()
	if err != nil {
		log.Warn("open terminal failed")
		terminalWidth = 0
	}
	if terminalWidth > terminalCharsetMaxCount {
		terminalWidth = terminalCharsetMaxCount
	}

	// init command
	shortDesc := fmt.Sprintf(`kplayer for golang %s Copyright(c) %s the ByteLang Studio (%s)
  libkplayer version: %s plugin version: %s license version: %s 
  toolchains %s C++ Standard %s on %s
  build with build-chains %s type with %s
  Hope you have a good experience.
`,
		types.MAJOR_TAG,
		info.Copyright,
		types.WebSite,
		info.MajorVersion,
		info.PluginVersion,
		info.LicenseVersion,
		info.ToolChains,
		info.CppStd,
		info.ArchiveVersion,
		info.BuildChains,
		info.BuildType,
	)

	for i := 0; uint(i) < terminalWidth; i++ {
		shortDesc = shortDesc + "-"
	}
	fmt.Println(shortDesc)

	rootCmd := &cobra.Command{
		Use: app.AppName,
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
	app.AddModuleCommands(rootCmd)
}

func messageConsumer(message *kpproto.KPMessage) {
	log.WithFields(log.Fields{"action": kpproto.EventAction_name[int32(message.Action)]}).Debug("receive broadcast message")

	var copyMsg kpproto.KPMessage
	for _, item := range app.ModuleManager.Modules {
		copyMsg = *message
		item.ParseMessage(&copyMsg)

		copyMsg = *message
		item.TriggerMessage(&copyMsg)
	}

	go func() {
		for _, item := range subscribeCollector {
			item <- *message
		}
	}()
}

func SubscribeMessage(name string) (chan kpproto.KPMessage, error) {
	subscribeMutex.Lock()
	defer subscribeMutex.Unlock()

	if _, ok := subscribeCollector[name]; ok {
		return nil, fmt.Errorf("subscribe name has been registed")
	}

	subscribeCollector[name] = make(chan kpproto.KPMessage, 500)
	return subscribeCollector[name], nil
}
