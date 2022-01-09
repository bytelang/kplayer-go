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
		Use:   "add",
		Short: "add output",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			var clientCtx *kptypes.ClientContext
			if ptr, err := kptypes.GetCommandContext(cmd, kptypes.ClientContextKey); err != nil {
				log.Fatal(err)
			} else {
				clientCtx = ptr.(*kptypes.ClientContext)
			}

			reply := &kpserver.OutputListReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Output.List", &kpserver.OutputListArgs{}, reply); err != nil {
				log.Error(err)
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
