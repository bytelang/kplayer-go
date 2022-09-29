package provider

import (
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	moduletypes "github.com/bytelang/kplayer/types/module"
	"github.com/bytelang/kplayer/types/server"
	"sync"
)

const (
	ModuleName = "resource"
)

const (
	CannotRemoveCurrentResource ResourceError = "can not remove playing resource"
	ResourceNotFound            ResourceError = "resource not found"
	ResourceUniqueHasExisted    ResourceError = "resource unique name has existed"
	ResourcePathCanNotBeEmpty   ResourceError = "resource path can not be empty"
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

	if len(resource.Path) == 0 {
		return ResourcePathCanNotBeEmpty
	}

	rs.resources = append(rs.resources, resource)
	return nil
}

// CalcMixResourceGroupPrimaryPath
// Under mixed resources, gets which resource should be selected as the primary resource
func CalcMixResourceGroupPrimaryPath(groups []*moduletypes.MixResourceGroup) (firstVideoResourceGroup *moduletypes.MixResourceGroup, firstAudioResourceGroup *moduletypes.MixResourceGroup, primaryResourceGroup *moduletypes.MixResourceGroup) {
	var firstVideoResource *moduletypes.MixResourceGroup = nil
	var firstAudioResource *moduletypes.MixResourceGroup = nil
	for _, groupItem := range groups {
		if groupItem.MediaType == moduletypes.ResourceMediaType_video && firstVideoResource == nil {
			firstVideoResource = groupItem
		}
		if groupItem.MediaType == moduletypes.ResourceMediaType_audio && firstAudioResource == nil {
			firstAudioResource = groupItem
		}

		// add groups
		mediaType := moduletypes.ResourceMediaType_video
		if groupItem.MediaType == moduletypes.ResourceMediaType_audio {
			mediaType = moduletypes.ResourceMediaType_audio
		}
		groups = append(groups, &moduletypes.MixResourceGroup{
			Path:           groupItem.Path,
			MediaType:      mediaType,
			PersistentLoop: groupItem.PersistentLoop,
		})
	}

	// calc primary resource
	var primaryResource *moduletypes.MixResourceGroup = firstVideoResource
	if primaryResource.PersistentLoop && !firstAudioResource.PersistentLoop {
		primaryResource = firstAudioResource
	}

	return firstVideoResource, firstAudioResource, primaryResource
}

func TransferConfigToModuleResourceGroup(configResourceGroups []*config.MixResourceGroup) []*moduletypes.MixResourceGroup {
	var groups []*moduletypes.MixResourceGroup

	for _, item := range configResourceGroups {
		mediaType := moduletypes.ResourceMediaType_video
		if item.MediaType == config.ResourceMediaType_audio {
			mediaType = moduletypes.ResourceMediaType_audio
		}
		group := &moduletypes.MixResourceGroup{
			Path:           item.Path,
			MediaType:      mediaType,
			PersistentLoop: item.PersistentLoop,
		}

		groups = append(groups, group)
	}

	return groups
}

func TransferServerToModuleResourceGroup(serverResourceGroups []*server.MixResourceGroup) []*moduletypes.MixResourceGroup {
	var groups []*moduletypes.MixResourceGroup

	for _, item := range serverResourceGroups {
		mediaType := moduletypes.ResourceMediaType_video
		if item.MediaType == server.ResourceMediaType_audio {
			mediaType = moduletypes.ResourceMediaType_audio
		}
		group := &moduletypes.MixResourceGroup{
			Path:           item.Path,
			MediaType:      mediaType,
			PersistentLoop: item.PersistentLoop,
		}

		groups = append(groups, group)
	}

	return groups
}

func GetResourceUniqueName(uniqueName string, path string, append ...string) string {
	if len(uniqueName) != 0 {
		return uniqueName
	}

	return kptypes.GetUniqueString(path, append...)
}
