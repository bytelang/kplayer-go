package provider

import (
    "github.com/bytelang/kplayer/client"
    "github.com/bytelang/kplayer/cmd/module/player/types"
)

func InitConfig(ctx client.ClientContext, provider Provider, config types.Config) {
    provider.SetConfig(config)
}
