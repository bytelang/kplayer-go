package provider

import (
	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	svrproto "github.com/bytelang/kplayer/types/server"
	log "github.com/sirupsen/logrus"
	"strings"
	"time"
)

type ProviderI interface {
	GetStartPoint() uint32
	GetPlayModel() config.PLAY_MODEL
	GetRPCParams() config.Rpc
	PlayStop(args *svrproto.PlayStopArgs) (*svrproto.PlayStopReply, error)
	PlayPause(args *svrproto.PlayPauseArgs) (*svrproto.PlayPauseReply, error)
	PlaySkip(args *svrproto.PlaySkipArgs) (*svrproto.PlaySkipReply, error)
	PlayContinue(args *svrproto.PlayContinueArgs) (*svrproto.PlayContinueReply, error)
	PlayDuration(args *svrproto.PlayDurationArgs) (*svrproto.PlayDurationReply, error)
	PlayInformation(args *svrproto.PlayInformationArgs) (*svrproto.PlayInformationReply, error)
}

var _ ProviderI = &Provider{}

// Provider play module provider
type Provider struct {
	module.ModuleKeeper

	// config
	startPoint uint32
	playMode   config.PLAY_MODEL
	rpc        config.Rpc

	// module member
	startTime time.Time

	// empty resource list for generate cache only
	GenerateCacheFlag bool
}

// NewProvider return provider
func NewProvider() *Provider {
	return &Provider{}
}

// InitConfig set module config on kplayer started
func (p *Provider) InitModule(ctx *kptypes.ClientContext, cfg *config.Play) {
	// set provider attribute
	p.startPoint = cfg.StartPoint
	playModel, ok := config.PLAY_MODEL_value[strings.ToUpper(cfg.PlayModel)]
	if !ok {
		log.Fatal("play model invalid")
	}
	p.playMode = config.PLAY_MODEL(playModel)

	p.rpc = *cfg.Rpc
}

func (p *Provider) ParseMessage(message *kpproto.KPMessage) {
	switch message.Action {
	case kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED:
		log.Info("kplayer start success")
		p.startTime = time.Now()
	}
}

func (p *Provider) ValidateConfig() error {
	return nil
}

func (p *Provider) GetStartPoint() uint32 {
	return p.startPoint
}

func (p *Provider) GetPlayModel() config.PLAY_MODEL {
	if p.GenerateCacheFlag {
		return config.PLAY_MODEL_LIST
	}

	return p.playMode
}
