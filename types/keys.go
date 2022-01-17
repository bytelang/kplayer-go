package types

import (
	"strconv"
)

var (
	MAJOR_TAG  string = "<MAJOR_TAG>"
	MAJOR_HASH string = "<MAJOR_TAG>"
	WebSite    string = "<WEB_SITE>"
)

const (
	FlagHome           = "home"
	FlagConfigFileName = "config"
	FlagLogLevel       = "log_level"
	FlagLogFormat      = "plain"
)

const (
	KplayerPluginSignHeader = "kplayer_signature"
)

const (
	DefaultRPCAddress string = "127.0.0.1"
	DefaultRPCPort    uint32 = 4156
)

// ErrorCode contains the exit code for server exit.
type ErrorCode struct {
	Code int
}

func (e ErrorCode) Error() string {
	return strconv.Itoa(e.Code)
}
