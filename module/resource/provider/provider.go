package provider

import (
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
    "sync"
    "time"
)

type ProviderI interface {
    ResourceAdd(resource *svrproto.ResourceAddArgs) (*svrproto.ResourceAddReply, error)
    ResourceRemove(resource *svrproto.ResourceRemoveArgs) (*svrproto.ResourceRemoveReply, error)
    ResourceList(*svrproto.ResourceListArgs) (*svrproto.ResourceListReply, error)
    ResourceAllList(*svrproto.ResourceAllListArgs) (*svrproto.ResourceAllListReply, error)
}

type Provider struct {
    module.ModuleKeeper

    // load config
    config config.Resource

    // module provider
    playProvider playprovider.ProviderI

    // module member
    currentIndex uint32
    inputs       []moduletypes.Resource
    input_mutex  sync.Mutex
}

func NewProvider(playProvider playprovider.ProviderI) *Provider {
    return &Provider{
        playProvider: playProvider,
    }
}

func (p *Provider) SetConfig(config config.Resource) {
    p.config = config
}

func (p *Provider) InitModule(ctx *kptypes.ClientContext, config config.Resource) {
    p.SetConfig(config)
    p.currentIndex = p.playProvider.GetStartPoint() - 1

    // initialize current index
    if p.currentIndex < 0 || p.currentIndex > uint32(len(p.config.Lists)) {
        p.currentIndex = 0
    }

    // initialize inputs
    for _, item := range p.config.Lists {
        p.inputs = append(p.inputs, moduletypes.Resource{
            Path:       item,
            Unique:     kptypes.GetRandString(6),
            CreateTime: uint64(time.Now().Unix()),
        })
    }
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
    switch message.Action {
    case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
        if len(p.inputs) == 0 {
            log.Info("the resource list is empty. waiting to add a resource")
            break
        }
        p.addNextResourceToCore()
    case kpproto.EVENT_MESSAGE_ACTION_RESOURCE_START:
        msg := &kpmsg.EventMessageResourceStart{}
        kptypes.UnmarshalProtoMessage(message.Body, msg)
        log.WithFields(log.Fields{"path": string(msg.Resource.Path)}).Info("start play resource")
    case kpproto.EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
        msg := &kpmsg.EventMessageResourceFinish{}
        kptypes.UnmarshalProtoMessage(message.Body, msg)
        if msg.Error != nil {
            log.WithFields(log.Fields{"error": string(msg.Error)}).Warn("play resource failed")
        } else {
            log.WithFields(log.Fields{"path": string(msg.Resource.Path), "index": p.currentIndex}).Info("finish play resource")
        }

        p.currentIndex = p.currentIndex + 1
        if p.currentIndex >= uint32(len(p.inputs)) {
            if p.playProvider.GetPlayModel() != config.PLAY_MODEL_name[int32(config.PLAY_MODEL_LOOP)] {
                stopCorePlay()
                return
            }
            p.currentIndex = 0
        }
        p.addNextResourceToCore()
    }
}

func (p *Provider) ValidateConfig() error {
    return nil
}

func (p *Provider) addNextResourceToCore() {
    if err := core.GetLibKplayerInstance().SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_ADD, &prompt.EventPromptResourceAdd{
        Resource: &kpproto.PromptResource{
            Path:   []byte(p.inputs[p.currentIndex].Path),
            Unique: []byte(p.inputs[p.currentIndex].Unique),
        },
    }); err != nil {
        log.Warn(err)
    }
}

func stopCorePlay() {
    if err := core.GetLibKplayerInstance().SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLAYER_STOP, &prompt.EventPromptPlayerStop{
    }); err != nil {
        log.Warn(err)
    }
}
