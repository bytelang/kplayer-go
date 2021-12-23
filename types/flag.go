package types

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path"
)

func GetHome(cmd *cobra.Command) (string, error) {
	home, err := cmd.Flags().GetString(FlagHome)
	return path.Join(home), err
}

func GetConfigFileName(cmd *cobra.Command) (string, error) {
	return cmd.Flags().GetString(FlagConfigFileName)
}

func GetLogLevel(cmd *cobra.Command) (log.Level, error) {
	level, err := cmd.Flags().GetString(FlagLogLevel)
	if err != nil {
		return log.PanicLevel, err
	}

	return log.ParseLevel(level)
}
