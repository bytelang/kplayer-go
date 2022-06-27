package provider

import (
	"context"
	"fmt"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/client"
	kpserver "github.com/bytelang/kplayer/types/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ModuleName,
		Short: "Output category",
		Long:  `Kplayer output management commands. control kplayer output add,remove...`,
	}
	cmd.AddCommand(addCommand())
	cmd.AddCommand(removeCommand())
	cmd.AddCommand(listCommand())

	return cmd
}

func addCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <output_path> [unique_name]",
		Short: `add output resource.`,
		Long: `output_path:
    support file rtmp ftp protocol
unique_name:
	optional argument. nickname for the output`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// args
			var path, unique string

			path = args[0]
			if len(args) > 1 {
				unique = args[1]
			} else {
				unique = kptypes.GetUniqueString(path)
			}

			// request
			conn, err := client.GrpcClientRequest(clientCtx.Config.Play.Rpc)
			if err != nil {
				return err
			}

			outputClient := kpserver.NewOutputGreeterClient(conn)
			reply, err := outputClient.OutputAdd(context.Background(), &kpserver.OutputAddArgs{
				Path:   path,
				Unique: unique,
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

func removeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <unique_name>",
		Short: `remove output resource by unique name. `,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// args
			uniqueName := args[0]

			// request
			conn, err := client.GrpcClientRequest(clientCtx.Config.Play.Rpc)
			if err != nil {
				return err
			}

			outputClient := kpserver.NewOutputGreeterClient(conn)
			reply, err := outputClient.OutputRemove(context.Background(), &kpserver.OutputRemoveArgs{
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

func listCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list output",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// request
			conn, err := client.GrpcClientRequest(clientCtx.Config.Play.Rpc)
			if err != nil {
				return err
			}

			outputClient := kpserver.NewOutputGreeterClient(conn)
			reply, err := outputClient.OutputList(context.Background(), &kpserver.OutputListArgs{})
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
