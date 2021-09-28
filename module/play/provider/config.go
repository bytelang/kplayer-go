package provider

import (
    "github.com/bytelang/kplayer/module/play/types"
    types2 "github.com/bytelang/kplayer/types"
)

func InitConfig(ctx types2.ClientContext, provider Provider, config types.Config) {
    provider.SetConfig(config)
}
