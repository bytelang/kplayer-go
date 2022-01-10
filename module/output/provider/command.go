package provider

import (
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
	cmd.AddCommand(AddCommand())
	cmd.AddCommand(ListCommand())

	return cmd
}

func AddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [output_path] [/unique_name]",
		Short: `add output resource. support file rtmp ftp protocol`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// args
			var path, unique string
			path = args[0]
			if len(args) > 1 {
				unique = args[1]
			}

			reply := &kpserver.OutputAddReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Output.Add", &kpserver.OutputAddArgs{
				Output: kpserver.Output{
					Path:   path,
					Unique: unique,
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

	return cmd
}

func ListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "list output",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.OutputListReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Output.List", &kpserver.OutputListArgs{}, reply); err != nil {
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
