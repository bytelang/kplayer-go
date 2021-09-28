package main

import (
    "context"
    "github.com/bytelang/kplayer/client"
    "github.com/bytelang/kplayer/cmd"
    "github.com/bytelang/kplayer/types"
    "github.com/rs/zerolog"
    log "github.com/sirupsen/logrus"
    "github.com/spf13/cobra"
    viper2 "github.com/spf13/viper"
    "os"
)

func init() {
    log.SetOutput(os.Stdout)
    log.SetReportCaller(true)
    log.SetLevel(log.TraceLevel)
}

func main() {
    rootCmd := cmd.NewRootCmd()

    if err := Execute(rootCmd, cmd.DefaultConfigFilePath); err != nil {
        switch e := err.(type) {
        case types.ErrorCode:
            os.Exit(e.Code)
        default:
            os.Exit(1)
        }
    }
}

func Execute(rootCmd *cobra.Command, defaultHome string) error {
    ctx := context.Background()
    ctx = context.WithValue(ctx, client.CommandContextKey, &client.ClientContext{})

    rootCmd.PersistentFlags().String("log_level", zerolog.InfoLevel.String(), "The logging level (trace|debug|info|warn|error|fatal|panic)")
    rootCmd.PersistentFlags().String("log_format", "plain", "The logging format (json|plain)")
    rootCmd.PersistentFlags().StringP("home", "", defaultHome, "directory for config and data")
    rootCmd.PersistentFlags().Bool("trace", false, "print out full stack trace on errors")
    rootCmd.PersistentPreRunE = types.ConcatCobraCmdFuncs(types.BindFlagsLoadViper, rootCmd.PersistentPreRunE)

    clientCtx := &client.ClientContext{
        Output:        os.Stdout,
        Viper:         viper2.New(),
        Config:        client.Config{},
    }

    return client.SetClientContextAndExecute(rootCmd, clientCtx)
}
