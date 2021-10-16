package types

import (
    "strconv"
)

const (
    FlagHome           = "home"
    FlagConfigFileName = "config"
)

// ErrorCode contains the exit code for server exit.
type ErrorCode struct {
    Code int
}

func (e ErrorCode) Error() string {
    return strconv.Itoa(e.Code)
}
