package app

import (
    "bytes"
    "encoding/json"
    "github.com/bytelang/kplayer/module"
    outputm "github.com/bytelang/kplayer/module/output"
    playm "github.com/bytelang/kplayer/module/play"
    resourcem "github.com/bytelang/kplayer/module/resource"
    "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    "github.com/spf13/cobra"
    "io/ioutil"
)

const (
    AppName               = "kplayer"
    DefaultConfigFileName = "config.json"
    DefaultConfigFilePath = "./"
    ConfigVersion         = "2.0.0"
)

var (
    ModuleManager = newModuleManager()
)

func newModuleManager() module.ModuleManager {
    playProvider := playm.NewAppModule()
    outputProvider := outputm.NewAppModule()
    resourceProvider := resourcem.NewAppModule(playProvider.GetConfig())
    return module.NewModuleManager(
        playProvider, outputProvider, resourceProvider,
    )
}

func AddCommands(rootCmd *cobra.Command) {
    for _, m := range ModuleManager {
        if cmd := m.GetCommand(); cmd != nil {
            rootCmd.AddCommand(cmd)
        }
    }
}

func AddInitCommands() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "init",
        Short: "init config file",
    }
    cmd.AddCommand(AddInitDefaultCommands())
    cmd.AddCommand(AddInitInteractionCommands())

    return cmd
}

func AddInitDefaultCommands() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "default",
        Short: "export default config file",
        RunE: func(cmd *cobra.Command, args []string) error {
            home, err := GetHome(cmd)
            if err != nil {
                return err
            }

            // init config
            cfg := getDefaultConfig()

            // export file
            return exportConfigFile(cfg, home+DefaultConfigFileName)
        },
    }

    return cmd
}

func AddInitInteractionCommands() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "interaction",
        Short: "interaction init config file",
        RunE: func(cmd *cobra.Command, args []string) error {
            home, err := cmd.Flags().GetString(types.FlagHome)
            if err != nil {
                return err
            }
            if home[:1] != "/" {
                home = home + "/"
            }

            // interaction
            cfg, err := initInteractionConfig()
            if err != nil {
                return err
            }

            // export file
            return exportConfigFile(cfg, home+DefaultConfigFileName)
        },
    }

    return cmd
}

func exportConfigFile(cfg *config.KPConfig, path string) error {
    d, err := json.Marshal(cfg)
    if err != nil {
        return err
    }

    var indentCfg bytes.Buffer
    if err := json.Indent(&indentCfg, d, "", "    "); err != nil {
        return err
    }

    return ioutil.WriteFile(path, indentCfg.Bytes(), 0666)
}
