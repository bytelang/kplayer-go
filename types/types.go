package types

import (
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "path/filepath"
    "strconv"
)

const (
    FlagHome   = "home"
    FlagConfig = "config"
)

// ErrorCode contains the exit code for server exit.
type ErrorCode struct {
    Code int
}

func (e ErrorCode) Error() string {
    return strconv.Itoa(e.Code)
}

type cobraCmdFunc func(cmd *cobra.Command, args []string) error

// Returns a single function that calls each argument function in sequence
// RunE, PreRunE, PersistentPreRunE, etc. all have this same signature
func ConcatCobraCmdFuncs(fs ...cobraCmdFunc) cobraCmdFunc {
    return func(cmd *cobra.Command, args []string) error {
        for _, f := range fs {
            if f != nil {
                if err := f(cmd, args); err != nil {
                    return err
                }
            }
        }
        return nil
    }
}

// Bind all flags and read the config into viper
func BindFlagsLoadViper(cmd *cobra.Command, args []string) error {
    // cmd.Flags() includes flags from this command and all persistent flags from the parent
    if err := viper.BindPFlags(cmd.Flags()); err != nil {
        return err
    }

    homeDir := viper.GetString(FlagHome)
    viper.Set(FlagHome, homeDir)
    viper.SetConfigName(FlagConfig)                         // name of config file (without extension)
    viper.AddConfigPath(homeDir)                            // search root directory
    viper.AddConfigPath(filepath.Join(homeDir, FlagConfig)) // search root directory /config

    // If a config file is found, read it in.
    if err := viper.ReadInConfig(); err == nil {
        // stderr, so if we redirect output to json file, this doesn't appear
        // fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
    } else if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
        // ignore not found error, return other errors
        return err
    }
    return nil
}
