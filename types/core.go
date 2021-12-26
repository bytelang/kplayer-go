package types

import (
	"github.com/bytelang/kplayer/core"
	"strings"
)

// GetCorePluginVersion
func GetCorePluginVersion() string {
	coreKplayer := core.GetLibKplayerInstance()
	version := coreKplayer.GetInformation().PluginVersion
	version = strings.ReplaceAll(version, ".", "")
	return version
}
