package provider

import (
	"github.com/spf13/cobra"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   ModuleName,
		Short: "plugin category",
		Long:  `Kplayer plugin management commands. control kplayer plugin add,remove...`,
	}
	cmd.AddCommand(AddCommand())

	return cmd
}

func AddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "add plugin",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
