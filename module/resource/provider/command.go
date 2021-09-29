package provider

import (
    "github.com/bytelang/kplayer/module/resource/types"
    "github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   types.ModuleName,
        Short: "resource category",
        Long:  `Kplayer resource management commands. control kplayer resource add,remove...`,
    }
    cmd.AddCommand(AddCommand())

    return cmd
}

func AddCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "add",
        Short: "add resource to playlist",
        RunE: func(cmd *cobra.Command, args []string) error {
            return nil
        },
    }

    return cmd
}
