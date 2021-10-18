package types

import (
    "strconv"
)

const (
    FlagHome           = "home"
    FlagConfigFileName = "config"
    FlagLogLevel       = "log_level"
    FlagLogFormat      = "plain"
)

// ErrorCode contains the exit code for server exit.
type ErrorCode struct {
    Code int
}

func (e ErrorCode) Error() string {
    return strconv.Itoa(e.Code)
}
