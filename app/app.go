package app

import (
	"github.com/bytelang/kplayer/module"
	outputm "github.com/bytelang/kplayer/module/output"
	playm "github.com/bytelang/kplayer/module/play"
	pluginm "github.com/bytelang/kplayer/module/plugin"
	resourcem "github.com/bytelang/kplayer/module/resource"
	"github.com/spf13/cobra"
)

const (
	AppName               = "kplayer"
	DefaultConfigFileName = "config"
	DefaultConfigFilePath = "./"
	ConfigVersion         = "2.0.0"
)

var (
	ModuleManager = newModuleManager()
)

func newModuleManager() module.ModuleManager {
	playProvider := playm.NewAppModule()
	outputProvider := outputm.NewAppModule()
	resourceProvider := resourcem.NewAppModule(playProvider)
	pluginProvider := pluginm.NewAppModule()
	mm := module.NewModuleManager(
		playProvider, outputProvider, resourceProvider, pluginProvider,
	)

	mm.SetOrderInitConfig(
		playProvider.GetModuleName(),
		outputProvider.GetModuleName(),
		resourceProvider.GetModuleName(),
		pluginProvider.GetModuleName(),
	)

	return mm
}

func AddModuleCommands(rootCmd *cobra.Command) {
	for _, m := range ModuleManager.Modules {
		if cmd := m.GetCommand(); cmd != nil {
			rootCmd.AddCommand(cmd)
		}
	}
}

func AddInitCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Init config file",
	}
	cmd.AddCommand(addInitDefaultCommands())
	cmd.AddCommand(addInitInteractionCommands())

	return cmd
}
