package provider

import (
	moduletypes "github.com/bytelang/kplayer/types/module"
	"sync"
)

const (
	ModuleName = "output"
)

const (
	OutputUniqueNotFound   OutputError = "output not found"
	OutputUniqueHasExisted OutputError = "output unique name has existed"
)

type OutputError string

func (e OutputError) Error() string {
	return string(e)
}

type Outputs struct {
	outputs []moduletypes.Output
	lock    sync.Mutex
}

func (o *Outputs) GetOutputByUnique(unique string) (*moduletypes.Output, int, error) {
	for key, item := range o.outputs {
		if item.Unique == unique {
			return &o.outputs[key], key, nil
		}
	}

	return nil, 0, OutputUniqueNotFound
}

func (o *Outputs) Exist(unique string) bool {
	for _, item := range o.outputs {
		if item.Unique == unique {
			return true
		}
	}

	return false
}

func (o *Outputs) RemoveOutputByUnique(unique string) (*moduletypes.Output, error) {
	o.lock.Lock()
	defer o.lock.Unlock()

	res, index, err := o.GetOutputByUnique(unique)
	if err != nil {
		return nil, err
	}

	var newOutput []moduletypes.Output
	newOutput = append(newOutput, o.outputs[:index]...)
	newOutput = append(newOutput, o.outputs[index+1:]...)

	o.outputs = newOutput

	return res, nil
}

func (o *Outputs) AppendOutput(output moduletypes.Output) error {
	o.lock.Lock()
	defer o.lock.Unlock()

	res, _, _ := o.GetOutputByUnique(output.Unique)
	if res != nil {
		return OutputUniqueHasExisted
	}

	o.outputs = append(o.outputs, output)
	return nil
}
