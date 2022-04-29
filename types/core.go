package types

import (
	"github.com/bytelang/kplayer/core"
	"strings"
)

// GetCorePluginVersion
func GetCorePluginVersion() string {
	coreKplayer := core.GetLibKplayerInstance()
	version := coreKplayer.GetInformation().PluginVersion
	versionArr := strings.Split(version, ".")
	for key, item := range versionArr {
		if len(item) == 1 && key != 0 {
			versionArr[key] = "0" + item
		}
	}
	return strings.Join(versionArr, "")
}
