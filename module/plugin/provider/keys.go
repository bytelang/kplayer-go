package provider

import (
	"crypto/md5"
	"fmt"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/api"
	moduletypes "github.com/bytelang/kplayer/types/module"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path"
	"sync"
)

const (
	ModuleName = "plugin"
)

const (
	PluginUniqueNotFound ResourceError = "plugin not found"
	PluginUniqueHasExist ResourceError = "plugin unique has exist"
	PluginFileNotFound   ResourceError = "plugin file not found"
)

const (
	FlagParams = "param"
)

const (
	PluginExtensionName = ".kpe"
)

type ResourceError string

func (r ResourceError) Error() string {
	return string(r)
}

type Plugins struct {
	plugins []moduletypes.Plugin
	lock    sync.Mutex
}

func (p *Plugins) GetPluginByUnique(unique string) (*moduletypes.Plugin, int, error) {
	for key, item := range p.plugins {
		if item.Unique == unique {
			return &p.plugins[key], key, nil
		}
	}

	return nil, 0, PluginUniqueNotFound
}

func (p *Plugins) Exist(unique string) bool {
	for _, item := range p.plugins {
		if item.Unique == unique {
			return true
		}
	}

	return false
}

func (ps *Plugins) RemovePluginByUnique(unique string) (*moduletypes.Plugin, error) {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	res, index, err := ps.GetPluginByUnique(unique)
	if res == nil {
		return nil, err
	}

	var newPlugins []moduletypes.Plugin
	newPlugins = append(newPlugins, ps.plugins[:index]...)
	newPlugins = append(newPlugins, ps.plugins[index+1:]...)

	ps.plugins = newPlugins

	return res, nil
}

func (ps *Plugins) AppendPlugin(plugin moduletypes.Plugin) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	res, _, _ := ps.GetPluginByUnique(plugin.Unique)
	if res != nil {
		return PluginUniqueHasExist
	}

	ps.plugins = append(ps.plugins, plugin)
	return nil
}

func GetPluginPath(name string, homePath string) string {
	return path.Join(homePath, "plugin", name+PluginExtensionName)
}

func InitPluginFile(name string, filePath string) error {
	logField := log.WithFields(log.Fields{"name": name, "path": filePath})

	// download file
	logField.Debug("get plugin file config")
	resp, err := kptypes.GetPlugin(&api.PluginInformationRequest{
		Name:    name,
		Version: kptypes.GetCorePluginVersion(),
	})
	if err != nil {
		return err
	}

	logField = logField.WithField("md5", resp.Md5)
	// check file
	if VerifyMD5Hash(filePath, resp.Md5) {
		logField.Debug("verify plugin file valid")
		return nil
	}

	logField.Debug("verify plugin file not exist or invalid. downloading")

	// download file
	if err := kptypes.DownloadFile(resp.DownloadUrl, filePath); err != nil {
		return err
	}

	logField.Info("plugin download success")
	return nil
}

func InitResourceFile(resourceType string, resourceName string, filePath string) error {
	resp, err := kptypes.GetResource(&api.ResourceInformationRequest{
		Type: resourceType,
		Name: resourceName,
	})
	if err != nil {
		return err
	}

	// download resource
	if err := kptypes.DownloadFile(resp.DownloadUrl, filePath); err != nil {
		log.WithField("file_path", filePath).Error("download file failed")
		return err
	}

	return nil
}

func VerifyMD5Hash(filePath string, fileMD5 string) bool {
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false
	}

	openFile, err := os.OpenFile(filePath, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return false
	}
	defer openFile.Close()

	fileContent, err := ioutil.ReadAll(openFile)
	if err != nil {
		return false
	}

	calcFileMD5 := fmt.Sprintf("%x", md5.Sum(fileContent))

	return calcFileMD5 == fileMD5
}
