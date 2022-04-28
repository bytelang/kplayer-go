package provider

import (
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
	"net/url"
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
	ResourceAdd(resource *svrproto.ResourceAddArgs) (*svrproto.ResourceAddReply, error)
	ResourceRemove(resource *svrproto.ResourceRemoveArgs) (*svrproto.ResourceRemoveReply, error)
	ResourceList(*svrproto.ResourceListArgs) (*svrproto.ResourceListReply, error)
	ResourceAllList(*svrproto.ResourceAllListArgs) (*svrproto.ResourceAllListReply, error)
	ResourceCurrent(*svrproto.ResourceCurrentArgs) (*svrproto.ResourceCurrentReply, error)
	ResourceSeek(*svrproto.ResourceSeekArgs) (*svrproto.ResourceSeekReply, error)
}

var _ ProviderI = &Provider{}

type Provider struct {
	module.ModuleKeeper

	// module provider
	playProvider playprovider.ProviderI

	// module member
	currentIndex    uint32
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
	p.currentIndex = p.playProvider.GetStartPoint() - 1
	p.allowExtensions = cfg.Extensions

	for _, item := range cfg.Lists {
		// add resource directory
		if files, err := kptypes.GetDirectorFiles(item); err == nil {
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
					log.WithFields(log.Fields{"path": item, "error": err}).Error("add resource to playlist failed")
				}
			}
			continue
		}

		// add resource file
		if err := p.inputs.AppendResource(moduletypes.Resource{
			Path:       item,
			Unique:     kptypes.GetUniqueString(item),
			Seek:       0,
			End:        -1,
			CreateTime: uint64(time.Now().Unix()),
		}); err != nil {
			log.WithFields(log.Fields{"path": item, "error": err}).Error("add resource to playlist failed")
		}
	}

	if p.playProvider.GetPlayModel() == config.PLAY_MODEL_RANDOM {
		p.currentIndex = uint32(rand.Intn(len(p.inputs.resources)))
	}
}

func (p *Provider) ValidateConfig() error {
	if p.currentIndex < 0 {
		return fmt.Errorf("start point invalid. cannot less than 1")
	} else if p.currentIndex >= uint32(len(p.inputs.resources)) {
		return fmt.Errorf("start point invalid. cannot great than total resource")
	}

	existName := []string{}
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
	case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
		if len(p.inputs.resources) == 0 {
			log.Info("the resource list is empty. waiting to add a resource")
			break
		}

		p.input_mutex.Lock()
		defer p.input_mutex.Unlock()
		p.addNextResourceToCore()
	case kpproto.EVENT_MESSAGE_ACTION_RESOURCE_START:
		msg := &kpmsg.EventMessageResourceStart{}
		kptypes.UnmarshalProtoMessage(message.Body, msg)
		log.WithFields(log.Fields{"path": msg.Resource.Path, "unique": msg.Resource.Unique}).
			Info("start play resource")

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
	case kpproto.EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
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

		p.currentIndex = p.currentIndex + 1

		// play_model
		switch p.playProvider.GetPlayModel() {
		case config.PLAY_MODEL_LIST:
			if p.currentIndex >= uint32(len(p.inputs.resources)) {
				log.Info("the playlist has been play completed")
				stopCorePlay()
				return
			}
		case config.PLAY_MODEL_LOOP:
			if p.currentIndex >= uint32(len(p.inputs.resources)) {
				p.currentIndex = 0
				log.Infof("Running mode on [%s]. will a new loop will take place...", strings.ToLower(p.playProvider.GetPlayModel().String()))
			}
		case config.PLAY_MODEL_QUEUE:
			if p.currentIndex >= uint32(len(p.inputs.resources)) {
				log.Infof("Running mode on [%s]. wait for the resource file to be added...", strings.ToLower(p.playProvider.GetPlayModel().String()))
				return // wait for new resource
			}
		case config.PLAY_MODEL_RANDOM:
			// refresh list
			p.randomModeUniqueNameList = []string{}
			for _, item := range p.inputs.resources {
				if !kptypes.ArrayInString(p.randomModeUniqueNameHistory, item.Unique) {
					p.randomModeUniqueNameList = append(p.randomModeUniqueNameList, item.Unique)
				}
			}

			// random index
			p.randomModeUniqueNameHistory = append(p.randomModeUniqueNameHistory, p.inputs.resources[p.currentIndex].Unique)
			p.currentIndex = uint32(rand.Intn(len(p.randomModeUniqueNameList)))
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
	pathUrl, err := url.Parse(currentResource.Path)
	if err == nil {
		if pathUrl.Scheme == "http" || pathUrl.Scheme == "https" {
			pathUrl.Query().Encode()
			encodePath = pathUrl.String()
		}
	}

	if err := core.GetLibKplayerInstance().SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_ADD, &prompt.EventPromptResourceAdd{
		Resource: &kpproto.PromptResource{
			Path:   encodePath,
			Unique: currentResource.Unique,
			Seek:   currentResource.Seek,
			End:    currentResource.End,
		},
	}); err != nil {
		log.Warn(err)
	}
}

func stopCorePlay() {
	if err := core.GetLibKplayerInstance().SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLAYER_STOP, &prompt.EventPromptPlayerStop{}); err != nil {
		log.Warn(err)
	}
}
