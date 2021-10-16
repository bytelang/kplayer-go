package types

import (
    "context"
    "fmt"
    "github.com/bytelang/kplayer/types/config"
    "io"
    "os"

    "github.com/spf13/cobra"
    "github.com/spf13/viper"
)

type KplayerContextKey string

func (k KplayerContextKey) String() string {
    return "kplayer_ctx" + string(k)
}

const (
    ClientContextKey        KplayerContextKey = "client.context"
    ModuleManagerContextKey KplayerContextKey = "module.manager"
)

type ClientContext struct {
    Output io.Writer
    Viper  *viper.Viper
    Config *config.KPConfig
}

func DefaultClientContext() *ClientContext {
    return &ClientContext{
        Output: os.Stdout,
        Viper:  viper.New(),
        Config: &config.KPConfig{},
    }
}

func GetCommandContext(cmd *cobra.Command, key KplayerContextKey) (interface{}, error) {
    if v := cmd.Context().Value(key); v != nil {
        return v, nil
    }

    return nil, fmt.Errorf("get context failed")
}

func SetCommandContextAndExecute(cmd *cobra.Command, ctx context.Context) error {
    return cmd.ExecuteContext(ctx)
}
