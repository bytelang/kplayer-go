package provider

import (
    "github.com/bytelang/kplayer/module/resource/provider"
    "github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   provider.ModuleName,
        Short: "output category",
        Long:  `Kplayer output management commands. control kplayer output add,remove...`,
    }
    cmd.AddCommand(AddCommand())

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
