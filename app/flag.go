package app

import (
    "github.com/bytelang/kplayer/types"
    "github.com/spf13/cobra"
)

func GetHome(cmd *cobra.Command) (string, error) {
    home, err := cmd.Flags().GetString(types.FlagHome)
    if err != nil {
        return "", err
    }
    if home[:1] != "/" {
        home = home + "/"
    }

    return home, nil
}

func GetConfigFileName(cmd *cobra.Command) (string, error) {
    return cmd.Flags().GetString(types.FlagConfigFileName)
}
