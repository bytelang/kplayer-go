package provider

import (
	"fmt"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/client"
	kpserver "github.com/bytelang/kplayer/types/server"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"strconv"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ModuleName,
		Short: "Resource category",
		Long:  `Kplayer resource management commands. control kplayer resource add,remove...`,
	}
	cmd.AddCommand(AddCommand())
	cmd.AddCommand(RemoveCommand())
	cmd.AddCommand(ListCommand())
	cmd.AddCommand(AllCommand())
	cmd.AddCommand(CurrentCommand())
	cmd.AddCommand(SeekCommand())

	return cmd
}

func AddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <input_path> [unique] [seek] [end]",
		Short: "add resource to playlist",
		Long: `input_path:
    resource file path. support [file/rtmp/ftp] protocel
unique:
    optional argument. resource unique name
seek:
    optional argument. start seek seconds position
end:
    optional argument. end seek seconds position`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// args
			var err error
			var path, unique string
			var seek, end int64

			path = args[0]
			if len(args) > 1 {
				unique = args[1]
			} else {
				unique = kptypes.GetRandString(6)
			}

			if len(args) > 2 {
				seek, err = strconv.ParseInt(args[2], 10, 64)
				if err != nil {
					return err
				}
			}

			if len(args) > 3 {
				end, err = strconv.ParseInt(args[3], 10, 64)
				if err != nil {
					return err
				}
			}

			// send request
			reply := &kpserver.ResourceAddReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Resource.Add", &kpserver.ResourceAddArgs{
				Path:   path,
				Unique: unique,
				Seek:   seek,
				End:    end,
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

func RemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <unique>",
		Short: "remove resource to playlist by unique name",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// args
			uniqueName := args[0]

			reply := &kpserver.ResourceRemoveReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Resource.Remove", &kpserver.ResourceRemoveArgs{
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
		Short: "gets the list of unplayed resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.ResourceListReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Resource.List", &kpserver.ResourceListArgs{}, reply); err != nil {
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

func AllCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "gets the list of resources",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.ResourceAllListReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Resource.AllList", &kpserver.ResourceAllListArgs{}, reply); err != nil {
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

func CurrentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "current",
		Short: "get the list of current play resource",
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			reply := &kpserver.ResourceCurrentReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Resource.Current", &kpserver.ResourceCurrentArgs{}, reply); err != nil {
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

func SeekCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "seek <unique> <seek>",
		Short: `seeks in current resource to seconds position`,
		Long: `unique:
    resource unique name
seek:
	seek seconds position`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// get client ctx
			clientCtx := kptypes.GetClientContextFromCommand(cmd)

			// args
			uniqueName := args[0]
			seekPosition, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return err
			}

			reply := &kpserver.ResourceSeekReply{}
			if err := client.ClientRequest(clientCtx.Config.Play.Rpc, "Resource.Seek", &kpserver.ResourceSeekArgs{
				Unique: uniqueName,
				Seek:   seekPosition,
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
