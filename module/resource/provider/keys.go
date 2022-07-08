package provider

import (
	moduletypes "github.com/bytelang/kplayer/types/module"
	"sync"
)

const (
	ModuleName = "resource"
)

const (
	CannotRemoveCurrentResource ResourceError = "can not remove playing resource"
	ResourceNotFound            ResourceError = "resource not found"
	ResourceUniqueHasExisted    ResourceError = "resource unique name has existed"
)

type ResourceError string

func (r ResourceError) Error() string {
	return string(r)
}

type Resources struct {
	resources []moduletypes.Resource
	lock      sync.Mutex
}

func (rs *Resources) GetResourceByUnique(unique string) (*moduletypes.Resource, int, error) {
	for key, item := range rs.resources {
		if item.Unique == unique {
			return &(rs.resources[key]), key, nil
		}
	}
	return nil, 0, ResourceNotFound
}

func (rs *Resources) GetResourceByIndex(index int) (*moduletypes.Resource, error) {
	if index > len(rs.resources) {
		return nil, ResourceNotFound
	}

	return &rs.resources[index], nil
}

func (rs *Resources) Exist(unique string) bool {
	for _, item := range rs.resources {
		if item.Unique == unique {
			return true
		}
	}

	return false
}

func (rs *Resources) RemoveResourceByUnique(unique string) (*moduletypes.Resource, int, error) {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	res, index, err := rs.GetResourceByUnique(unique)
	if err != nil {
		return nil, 0, err
	}

	var newResource []moduletypes.Resource
	newResource = append(newResource, (rs.resources)[:index]...)
	newResource = append(newResource, (rs.resources)[index+1:]...)

	rs.resources = newResource
	return res, index, nil
}

func (rs *Resources) AppendResource(resource moduletypes.Resource) error {
	rs.lock.Lock()
	defer rs.lock.Unlock()

	res, _, _ := rs.GetResourceByUnique(resource.Unique)
	if res != nil {
		return ResourceUniqueHasExisted
	}

	rs.resources = append(rs.resources, resource)
	return nil
}
