package provider

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	playprovider "github.com/bytelang/kplayer/module/play/provider"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	kpmsg "github.com/bytelang/kplayer/types/core/proto/msg"
	"github.com/bytelang/kplayer/types/core/proto/prompt"
	moduletypes "github.com/bytelang/kplayer/types/module"
	svrproto "github.com/bytelang/kplayer/types/server"
	log "github.com/sirupsen/logrus"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type ProviderI interface {
	ResourceAdd(context.Context, *svrproto.ResourceAddArgs) (*svrproto.ResourceAddReply, error)
	ResourceRemove(context.Context, *svrproto.ResourceRemoveArgs) (*svrproto.ResourceRemoveReply, error)
	ResourceList(context.Context, *svrproto.ResourceListArgs) (*svrproto.ResourceListReply, error)
	ResourceListAll(context.Context, *svrproto.ResourceListAllArgs) (*svrproto.ResourceListAllReply, error)
	ResourceCurrent(context.Context, *svrproto.ResourceCurrentArgs) (*svrproto.ResourceCurrentReply, error)
	ResourceSeek(context.Context, *svrproto.ResourceSeekArgs) (*svrproto.ResourceSeekReply, error)
}

var _ ProviderI = &Provider{}

type Provider struct {
	module.ModuleKeeper
	svrproto.UnimplementedResourceGreeterServer

	// module provider
	playProvider playprovider.ProviderI

	// module member
	currentIndex    int
	inputs          Resources
	allowExtensions []string

	// will reset seek attribute
	// set resource seek on replayed need set the resource attribute
	resetInputs map[string]int64

	input_mutex sync.Mutex

	// random history list
	randomModeUniqueNameList    []string
	randomModeUniqueNameHistory []string
}

var _ ProviderI = &Provider{}

func NewProvider(playProvider playprovider.ProviderI) *Provider {
	return &Provider{
		playProvider: playProvider,
		resetInputs:  make(map[string]int64),
	}
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, cfg *config.Resource) {
	// initialize attribute
	p.currentIndex = int(p.playProvider.GetStartPoint()) - 1
	p.allowExtensions = cfg.Extensions

	for _, item := range cfg.Lists {
		// parse resource item
		res, err := kptypes.GetResourceItemByAny(item)
		if err != nil {
			log.WithField("content", item.String()).Fatal("not in the expected format")
		}

		switch assertRes := res.(type) {
		case *config.SingleResource:
			// add resource directory
			if files, err := kptypes.GetDirectorFiles(assertRes.Path); err == nil {
				// sort file
				sort.Strings(files)

				for _, f := range files {
					ext := filepath.Ext(f)
					if len(ext) > 1 {
						ext = ext[1:]
					}
					if p.allowExtensions != nil && !kptypes.ArrayInString(p.allowExtensions, ext) {
						continue
					}

					if err := p.inputs.AppendResource(moduletypes.Resource{
						Path:       f,
						Unique:     kptypes.GetUniqueString(f),
						Seek:       0,
						End:        -1,
						CreateTime: uint64(time.Now().Unix()),
					}); err != nil {
						log.WithFields(log.Fields{"path": assertRes.Path, "error": err}).Error("add resource to playlist failed")
					}
				}
				continue
			}

			// add resource file
			uniqueName := assertRes.Unique
			if len(uniqueName) == 0 {
				uniqueName = kptypes.GetUniqueString(assertRes.Path)
			}

			if err := p.inputs.AppendResource(moduletypes.Resource{
				Path:       assertRes.Path,
				Unique:     uniqueName,
				Seek:       assertRes.Seek,
				End:        assertRes.End,
				CreateTime: uint64(time.Now().Unix()),
			}); err != nil {
				log.WithFields(log.Fields{"path": assertRes.Path, "error": err, "type": "single"}).Error("add resource to playlist failed")
			}
		case *config.MixResource:
			// add resource mix file
			var firstVideoResource *config.MixResourceGroup = nil
			var firstAudioResource *config.MixResourceGroup = nil

			var groups []*moduletypes.MixResourceGroup
			for _, groupItem := range assertRes.Groups {
				if groupItem.MediaType == config.ResourceMediaType_video && firstVideoResource == nil {
					firstVideoResource = groupItem
				}
				if groupItem.MediaType == config.ResourceMediaType_audio && firstAudioResource == nil {
					firstAudioResource = groupItem
				}

				// add groups
				mediaType := moduletypes.ResourceMediaType_video
				if groupItem.MediaType == config.ResourceMediaType_audio {
					mediaType = moduletypes.ResourceMediaType_audio
				}
				groups = append(groups, &moduletypes.MixResourceGroup{
					Path:           groupItem.Path,
					MediaType:      mediaType,
					PersistentLoop: groupItem.PersistentLoop,
				})
			}

			// calc primary resource
			var primaryResource *config.MixResourceGroup = firstVideoResource
			if primaryResource.PersistentLoop && !firstAudioResource.PersistentLoop {
				primaryResource = firstAudioResource
			}

			// eliminating all resources requires a loop
			if firstVideoResource.PersistentLoop && firstAudioResource.PersistentLoop {
				for key, _ := range groups {
					groups[key].PersistentLoop = false
				}
			}

			uniqueName := assertRes.Unique
			if len(uniqueName) == 0 {
				uniqueName = kptypes.GetUniqueString(primaryResource.Path, "MIX")
			}
			if err := p.inputs.AppendResource(moduletypes.Resource{
				Path:            primaryResource.Path,
				Unique:          uniqueName,
				Seek:            assertRes.Seek,
				End:             assertRes.End,
				CreateTime:      uint64(time.Now().Unix()),
				MixResourceType: true,
				Groups:          groups,
			}); err != nil {
				log.WithFields(log.Fields{"path": primaryResource, "groups": assertRes.Groups, "error": err, "type": "mix"}).Error("add resource to playlist failed")
			}
		default:
			log.WithField("error", "invalid resource type").Fatal(item)
		}
	}

	if p.playProvider.GetPlayModel() == config.PLAY_MODEL_RANDOM {
		p.currentIndex = rand.Intn(len(p.inputs.resources))
	}
}

func (p *Provider) ValidateConfig() error {
	if p.currentIndex < 0 {
		return fmt.Errorf("start point invalid. cannot less than 1")
	} else if p.currentIndex >= len(p.inputs.resources) {
		return fmt.Errorf("start point invalid. cannot great than total resource")
	}

	var existName []string
	for _, item := range p.inputs.resources {
		if kptypes.ArrayInString(existName, item.Unique) {
			return ResourceUniqueHasExisted
		}

		existName = append(existName, item.Unique)
	}

	return nil
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
	switch message.Action {
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_STARTED:
		if len(p.inputs.resources) == 0 {
			log.Info("the resource list is empty. waiting to add a resource")
			break
		}

		p.input_mutex.Lock()
		defer p.input_mutex.Unlock()
		p.addNextResourceToCore()
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_START:
		msg := &kpmsg.EventMessageResourceStart{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		log.WithFields(log.Fields{"path": msg.Resource.Path, "unique": msg.Resource.Unique}).
			Debug("start play resource")

		res, _, err := p.inputs.GetResourceByUnique(msg.Resource.Unique)
		if err != nil {
			log.WithFields(log.Fields{"unique": msg.Resource.Unique, "path": msg.Resource.Path}).Warn(err)
			break
		}

		res.StartTime = uint64(time.Now().Unix())
		res.EndTime = 0

		// reset resource seek attribute
		if seek, ok := p.resetInputs[msg.Resource.Unique]; ok {
			res.Seek = seek
		}
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_CHECKED:
		msg := &kpmsg.EventMessageResourceChecked{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)

		// log field single and mix
		logFields := log.Fields{"path": msg.Resource.Path,
			"unique":   msg.Resource.Unique,
			"type":     strings.ToLower(msg.Resource.InputType.String()),
			"duration": msg.InputAttribute.Duration}

		if p.playProvider.GetCacheOn() {
			logFields["hit_cache"] = msg.HitCache
		}
		log.WithFields(logFields).Info("checked play resource")
	case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
		msg := &kpmsg.EventMessageResourceFinish{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)

		logFields := log.WithFields(log.Fields{"unique": msg.Resource.Unique, "path": msg.Resource.Path})

		if len(msg.Error) != 0 {
			logFields.WithFields(log.Fields{"error": msg.Error}).Warn("play resource failed")
		} else {
			logFields.WithFields(log.Fields{"path": msg.Resource.Path, "unique": msg.Resource.Unique}).
				Info("finish play resource")
		}

		p.input_mutex.Lock()
		defer p.input_mutex.Unlock()

		// get resource
		res, _, err := p.inputs.GetResourceByUnique(msg.Resource.Unique)
		if err != nil {
			logFields.Warn(err)
			break
		}
		res.EndTime = uint64(time.Now().Unix())

		// play_model
		switch p.playProvider.GetPlayModel() {
		case config.PLAY_MODEL_LIST:
			p.currentIndex = p.currentIndex + 1
			if p.currentIndex >= len(p.inputs.resources) {
				log.Info("the playlist has been play completed")
				stopCorePlay()
				return
			}
		case config.PLAY_MODEL_LOOP:
			p.currentIndex = p.currentIndex + 1
			if p.currentIndex >= len(p.inputs.resources) {
				p.currentIndex = 0
				log.Infof("running mode on [%s]. will a new loop will take place...", strings.ToLower(p.playProvider.GetPlayModel().String()))
			}
		case config.PLAY_MODEL_QUEUE:
			p.currentIndex = p.currentIndex + 1
			if p.currentIndex >= len(p.inputs.resources) {
				log.Infof("running mode on [%s]. wait for the resource file to be added...", strings.ToLower(p.playProvider.GetPlayModel().String()))
				return // wait for new resource
			}
		case config.PLAY_MODEL_RANDOM:
			// refresh list
			p.randomModeUniqueNameList = []string{}
			for len(p.randomModeUniqueNameList) == 0 {
				for _, item := range p.inputs.resources {
					if !kptypes.ArrayInString(p.randomModeUniqueNameHistory, item.Unique) {
						p.randomModeUniqueNameList = append(p.randomModeUniqueNameList, item.Unique)
					}
				}

				if len(p.randomModeUniqueNameList) == 0 {
					p.randomModeUniqueNameHistory = []string{}
				}
			}

			// random index
			p.randomModeUniqueNameHistory = append(p.randomModeUniqueNameHistory, p.inputs.resources[p.currentIndex].Unique)
			p.currentIndex = rand.Intn(len(p.randomModeUniqueNameList))
		}
		p.addNextResourceToCore()
	}
}

func (p *Provider) addNextResourceToCore() {
	currentResource, err := p.inputs.GetResourceByIndex(p.currentIndex)
	if err != nil {
		log.Fatal("get resource failed")
		return
	}

	encodePath := currentResource.Path

	// protocol url encode
	encodePath = kptypes.PathUrlEncode(currentResource.Path)

	// input type
	inputType := kpproto.ResourceInputType_RESOURCE_INPUT_TYPE_SINGLE
	if currentResource.MixResourceType {
		inputType = kpproto.ResourceInputType_RESOURCE_INPUT_TYPE_MIX
	}

	// groups
	var groups []*kpproto.ResourceGroup
	for _, item := range currentResource.Groups {
		mediaType := kpproto.ResourceMediaType_RESOURCE_MEDIA_TYPE_VIDEO
		if item.MediaType == moduletypes.ResourceMediaType_audio {
			mediaType = kpproto.ResourceMediaType_RESOURCE_MEDIA_TYPE_AUDIO
		}
		groups = append(groups, &kpproto.ResourceGroup{
			Path:           kptypes.PathUrlEncode(item.Path),
			MediaType:      mediaType,
			PersistentLoop: item.PersistentLoop,
		})
	}

	if err := core.GetLibKplayerInstance().SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &prompt.EventPromptResourceAdd{
		Resource: &kpproto.PromptResource{
			Path:      encodePath,
			Unique:    currentResource.Unique,
			Seek:      currentResource.Seek,
			End:       currentResource.End,
			InputType: inputType,
			Groups:    groups,
		},
	}); err != nil {
		log.Warn(err)
	}
}

func stopCorePlay() {
	if err := core.GetLibKplayerInstance().SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &prompt.EventPromptPlayerStop{}); err != nil {
		log.Warn(err)
	}
}
