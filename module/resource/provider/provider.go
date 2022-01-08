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
	"strings"
	"sync"
	"time"
)

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
	currentIndex uint32
	inputs       Resources

	// will reset seek attribute
	// set resource seek on replayed need set the resource attribute
	resetInputs map[string]int64

	input_mutex sync.Mutex
}

var _ ProviderI = &Provider{}

func NewProvider(playProvider playprovider.ProviderI) *Provider {
	return &Provider{
		playProvider: playProvider,
		resetInputs:  make(map[string]int64),
	}
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config *config.Resource, homePath string) {
	// initialize attribute
	p.currentIndex = p.playProvider.GetStartPoint() - 1

	for _, item := range config.Lists {
		p.inputs = append(p.inputs, moduletypes.Resource{
			Path:       item,
			Unique:     kptypes.GetRandString(6),
			Seek:       0,
			End:        -1,
			CreateTime: uint64(time.Now().Unix()),
		})
	}
}

func (p *Provider) ValidateConfig() error {
	if p.currentIndex < 0 {
		return fmt.Errorf("start point invalid. cannot less than 1")
	} else if p.currentIndex > uint32(len(p.inputs)) {
		return fmt.Errorf("start point invalid. cannot great than total resource")
	}

	return nil
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
	switch message.Action {
	case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
		if len(p.inputs) == 0 {
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
		if len(msg.Error) != 0 {
			log.WithFields(log.Fields{"error": msg.Error}).Warn("play resource failed")
		} else {
			log.WithFields(log.Fields{"path": msg.Resource.Path, "unique": msg.Resource.Unique}).
				Info("finish play resource")
		}

		p.input_mutex.Lock()
		defer p.input_mutex.Unlock()

		// get resource
		res, _, err := p.inputs.GetResourceByUnique(msg.Resource.Unique)
		if err != nil {
			log.WithFields(log.Fields{"unique": msg.Resource.Unique, "path": msg.Resource.Path}).Warn(err)
			break
		}
		res.EndTime = uint64(time.Now().Unix())

		p.currentIndex = p.currentIndex + 1
		if p.currentIndex >= uint32(len(p.inputs)) {
			if p.playProvider.GetPlayModel() != strings.ToLower(config.PLAY_MODEL_name[int32(config.PLAY_MODEL_LOOP)]) {
				log.Info("the playlist has been play completed")
				stopCorePlay()
				return
			}
			p.currentIndex = 0
		}
		p.addNextResourceToCore()
	}
}

func (p *Provider) addNextResourceToCore() {
	if err := core.GetLibKplayerInstance().SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_ADD, &prompt.EventPromptResourceAdd{
		Resource: &kpproto.PromptResource{
			Path:   p.inputs[p.currentIndex].Path,
			Unique: p.inputs[p.currentIndex].Unique,
			Seek:   p.inputs[p.currentIndex].Seek,
			End:    p.inputs[p.currentIndex].End,
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
