package provider

import (
	"context"
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
			conn, err := client.GrpcClientRequest(clientCtx.Config.Play.Rpc)
			if err != nil {
				return err
			}

			pluginClient := kpserver.NewPluginGreeterClient(conn)
			reply, err := pluginClient.PluginAdd(context.Background(), &kpserver.PluginAddArgs{
				Path:   name,
				Unique: unique,
				Params: params,
			})
			if err != nil {
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

			conn, err := client.GrpcClientRequest(clientCtx.Config.Play.Rpc)
			if err != nil {
				return err
			}

			pluginClient := kpserver.NewPluginGreeterClient(conn)
			reply, err := pluginClient.PluginRemove(context.Background(), &kpserver.PluginRemoveArgs{
				Unique: uniqueName,
			})
			if err != nil {
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

			// request
			conn, err := client.GrpcClientRequest(clientCtx.Config.Play.Rpc)
			if err != nil {
				return err
			}

			pluginClient := kpserver.NewPluginGreeterClient(conn)
			reply, err := pluginClient.PluginList(context.Background(), &kpserver.PluginListArgs{})
			if err != nil {
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
			conn, err := client.GrpcClientRequest(clientCtx.Config.Play.Rpc)
			if err != nil {
				return err
			}

			pluginClient := kpserver.NewPluginGreeterClient(conn)
			reply, err := pluginClient.PluginUpdate(context.Background(), &kpserver.PluginUpdateArgs{
				Unique: uniqueName,
				Params: params,
			})
			if err != nil {
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
