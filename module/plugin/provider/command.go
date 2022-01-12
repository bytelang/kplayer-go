package provider

import (
	"fmt"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/client"
	kpserver "github.com/bytelang/kplayer/types/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strings"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ModuleName,
		Short: "Plugin category",
		Long:  `Kplayer plugin management commands. control kplayer plugin add,remove...`,
	}
	cmd.AddCommand(AddCommand())
	cmd.AddCommand(RemoveCommand())
	cmd.AddCommand(ListCommand())
	cmd.AddCommand(UpdateCommand())

	return cmd
}

func AddCommand() *cobra.Command {
	var paramsFlagValue []string

	cmd := &cobra.Command{
		Use:   "add <name> [unique]",
		Short: "add plugin",
		Long: `name:
    plugin name on your add plugin attribute
unique:
	optional argument. plugin nickname`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// args
			var name, unique string

			name = args[0]
			if len(args) > 1 {
				unique = args[1]
			} else {
				unique = kptypes.GetRandString(6)
			}

			params, err := parseFlagParams(paramsFlagValue)
			if err != nil {
				return err
			}

			// request
			reply := &kpserver.PluginAddReplay{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Plugin.Add", &kpserver.PluginAddReplay{
				Plugin: kpserver.Plugin{
					Path:   name,
					Unique: unique,
					Params: params,
				},
			}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&paramsFlagValue, FlagParams, "p", []string{}, "e.g: fontsize=19")

	return cmd
}

func RemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <unique>",
		Short: "remove plugin",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// args
			uniqueName := args[0]

			reply := &kpserver.PluginRemoveReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Plugin.Remove", &kpserver.PluginRemoveArgs{
				Unique: uniqueName,
			}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	return cmd
}

func ListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.PluginListReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Plugin.List", &kpserver.PluginListReply{
			}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	return cmd
}

func UpdateCommand() *cobra.Command {
	var paramsFlagValue []string

	cmd := &cobra.Command{
		Use:   "update <unique>",
		Short: "update plugin",
		Long: `unique:
    plugin unique name`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			uniqueName := args[0]

			params, err := parseFlagParams(paramsFlagValue)
			if err != nil {
				return err
			}

			// request
			reply := &kpserver.PluginUpdateReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Plugin.Update", &kpserver.PluginUpdateArgs{
				Unique: uniqueName,
				Params: params,
			}, reply); err != nil {
				log.Error(err)
				return nil
			}

			yaml, err := kptypes.FormatYamlProtoMessage(reply)
			if err != nil {
				return err
			}
			fmt.Print(yaml)

			return nil
		},
	}

	cmd.Flags().StringArrayVarP(&paramsFlagValue, FlagParams, "p", []string{}, "e.g: fontsize=19")

	return cmd
}

func parseFlagParams(paramsFlagValue []string) (map[string]string, error) {
	params := make(map[string]string)

	for _, item := range paramsFlagValue {
		splitArr := strings.Split(item, "=")
		if len(splitArr) < 2 {
			return nil, fmt.Errorf("params invalid. argument: %s", item)
		}
		params[splitArr[0]] = splitArr[1]
	}

	return params, nil
}
