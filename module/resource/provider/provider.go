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
    log "github.com/sirupsen/logrus"
)

type Provider struct {
    module.ModuleKeeper
    config     config.Resource
    playConfig playprovider.ProviderI

    currentIndex uint32
}

func NewProvider(playConfig playprovider.ProviderI) *Provider {
    return &Provider{
        playConfig: playConfig,
    }
}

func (p *Provider) SetConfig(config config.Resource) {
    p.config = config
}

func (p *Provider) InitModuleConfig(ctx *kptypes.ClientContext, config config.Resource) {
    p.SetConfig(config)
    p.currentIndex = p.playConfig.GetStartPoint() - 1

    if p.currentIndex < 0 || p.currentIndex > uint32(len(config.Lists)) {
        p.currentIndex = 0
    }
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
    switch message.Action {
    case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
        if len(p.config.Lists) == 0 {
            log.Info("the resource list is empty. waiting to add a resource")
            break
        }
        p.addNextResourceAdd()
    case kpproto.EVENT_MESSAGE_ACTION_RESOURCE_START:
        msg := &kpmsg.EventMessageResourceStart{}
        kptypes.UnmarshalProtoMessage(message.Body, msg)
        log.Info("start play resource: %s", string(msg.Resource.Path))
    case kpproto.EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
        msg := &kpmsg.EventMessageResourceFinish{}
        kptypes.UnmarshalProtoMessage(message.Body, msg)
        if msg.Error != nil {
            log.Warn("play resource failed: %s", string(msg.Error))
        } else {
            log.Info("finish play resource: %s; index: %d", string(msg.Resource.Path), p.currentIndex)
        }

        p.currentIndex = p.currentIndex + 1
        if p.currentIndex >= uint32(len(p.config.Lists)) {
            if p.playConfig.GetPlayModel() != config.PLAY_MODEL_name[int32(config.PLAY_MODEL_LOOP)] {
                stopCorePlay()
                return
            }
            p.currentIndex = 0
        }
        p.addNextResourceAdd()
    }

    p.Trigger(message)
}

func (p *Provider) ValidateConfig() error {
    return nil
}

func (p *Provider) addNextResourceAdd() {
    if err := core.GetLibKplayerInstance().SendPrompt(kpproto.EVENT_PROMPT_ACTION_RESOURCE_ADD, &prompt.EventPromptResourceAdd{
        Resource: &kpproto.PromptResource{
            Path:   []byte(p.config.Lists[p.currentIndex]),
            Unique: []byte(kptypes.GetRandString(6)),
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
