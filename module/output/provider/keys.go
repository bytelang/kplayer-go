package provider

import (
    moduletypes "github.com/bytelang/kplayer/types/module"
)

const (
    ModuleName = "output"
)

type Outputs []moduletypes.Output

func (o *Outputs) GetOutputByUnique(unique string) *moduletypes.Output {
    for key, item := range *o {
        if item.Unique == unique {
            return &(*o)[key]
        }
    }

    return nil
}
