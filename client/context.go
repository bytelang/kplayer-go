package client

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "io"
)

const (
    CommandContextKey = "command.context"
)

type ClientContext struct {
    Output io.Writer
    Viper  *viper.Viper
    Config Config
}

func GetClientContext(cmd *cobra.Command) (*ClientContext, error) {
    if v := cmd.Context().Value(CommandContextKey); v != nil {
        clientCtxPtr := v.(*ClientContext)
        return clientCtxPtr, nil
    }

    return nil, fmt.Errorf("get context failed.")
}

func SetClientContextAndExecute(cmd *cobra.Command, clientCtx *ClientContext) error {
    ctx := context.Background()
    ctx = context.WithValue(ctx, CommandContextKey, &ClientContext{})
    return cmd.ExecuteContext(ctx)
}
