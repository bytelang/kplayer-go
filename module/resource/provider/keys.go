package provider

import moduletypes "github.com/bytelang/kplayer/types/module"

const (
	ModuleName = "resource"
)

const (
	CannotRemoveCurrentResource ResourceError = "can not remove playing resource"
	ResourceNotFound            ResourceError = "resource not found"
)

type ResourceError string

func (r ResourceError) Error() string {
	return string(r)
}

type Resources []moduletypes.Resource

func (rs *Resources) GetResourceByUnique(unique string) (*moduletypes.Resource, int, error) {
	for key, item := range *rs {
		if item.Unique == unique {
			return &(*rs)[key], key, nil
		}
	}
	return nil, 0, ResourceNotFound
}

func (rs *Resources) RemoveResourceByUnique(unique string) (*moduletypes.Resource, error) {
	res, index, err := rs.GetResourceByUnique(unique)
	if res == nil {
		return nil, err
	}

	var newInputs Resources
	newInputs = append(newInputs, (*rs)[:index]...)
	newInputs = append(newInputs, (*rs)[index+1:]...)

	(*rs) = newInputs
	return res, nil
}
