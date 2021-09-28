package types

import (
    "context"
    "fmt"
    "github.com/spf13/cobra"
    "github.com/spf13/viper"
    "io"
    "os"
)

const (
    ClientContextKey = "client.context"
    AppContextKey    = "app.context"
)

type ClientContext struct {
    Output io.Writer
    Viper  *viper.Viper
    Config Config
}

func DefaultClientContext() *ClientContext {
    return &ClientContext{
        Output: os.Stdout,
        Viper:  viper.New(),
        Config: Config{},
    }
}

func GetCommandContext(cmd *cobra.Command, key string) (interface{}, error) {
    if v := cmd.Context().Value(key); v != nil {
        return v, nil
    }

    return nil, fmt.Errorf("get context failed.")
}

func SetCommandContextAndExecute(cmd *cobra.Command, ctx context.Context) error {
    return cmd.ExecuteContext(ctx)
}
