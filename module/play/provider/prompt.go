package provider

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/bytelang/kplayer/types/core/proto/msg"
	"github.com/bytelang/kplayer/types/core/proto/prompt"
	svrproto "github.com/bytelang/kplayer/types/server"
	"time"
)

func (p *Provider) PlayStop(ctx context.Context, args *svrproto.PlayStopArgs) (*svrproto.PlayStopReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &prompt.EventPromptPlayerStop{}); err != nil {
		return nil, err
	}

	// register prompt
	endedMsg := &msg.EventMessagePlayerEnded{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_STOP, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, endedMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(endedMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", endedMsg.Error)
	}

	return &svrproto.PlayStopReply{}, nil
}

func (p *Provider) PlayPause(ctx context.Context, args *svrproto.PlayPauseArgs) (*svrproto.PlayPauseReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_PAUSE, &prompt.EventPromptPlayerPause{}); err != nil {
		return nil, err
	}

	// register prompt
	pauseMsg := &msg.EventMessagePlayerPause{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_PAUSE, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, pauseMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(pauseMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", pauseMsg.Error)
	}

	return &svrproto.PlayPauseReply{}, nil
}

func (p *Provider) PlaySkip(ctx context.Context, args *svrproto.PlaySkipArgs) (*svrproto.PlaySkipReply, error) {
	// send skip prompt
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_SKIP, &prompt.EventPromptPlayerSkip{}); err != nil {
		return nil, err
	}

	// register prompt
	skipMsg := &msg.EventMessagePlayerSkip{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_SKIP, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, skipMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(skipMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", skipMsg.Error)
	}

	return &svrproto.PlaySkipReply{}, nil
}

func (p *Provider) PlayContinue(ctx context.Context, args *svrproto.PlayContinueArgs) (*svrproto.PlayContinueReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_CONTINUE, &prompt.EventPromptPlayerContinue{}); err != nil {
		return nil, err
	}

	// register prompt
	continueMsg := &msg.EventMessagePlayerContinue{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_CONTINUE, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, continueMsg)
		return true
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(continueMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", continueMsg.Error)
	}

	return &svrproto.PlayContinueReply{}, nil
}

func (p *Provider) PlayDuration(ctx context.Context, args *svrproto.PlayDurationArgs) (*svrproto.PlayDurationReply, error) {
	reply := &svrproto.PlayDurationReply{
		StartTimestamp:    uint64(p.startTime.Unix()),
		DurationTimestamp: uint64(time.Now().Unix() - p.startTime.Unix()),
	}
	return reply, nil
}

func (p *Provider) PlayInformation(ctx context.Context, args *svrproto.PlayInformationArgs) (*svrproto.PlayInformationReply, error) {
	coreKplayer := core.GetLibKplayerInstance()
	// get core information
	info := coreKplayer.GetInformation()

	reply := &svrproto.PlayInformationReply{
		MajorVersion:       kptypes.MAJOR_TAG,
		LibkplayerVersion:  info.MajorVersion,
		PluginVersion:      info.PluginVersion,
		LicenseVersion:     info.LicenseVersion,
		StartTime:          p.startTime.String(),
		StartTimeTimestamp: uint64(p.startTime.Unix()),
	}

	return reply, nil
}

func (p *Provider) GetRPCParams() config.Server {
	return p.rpc
}

func (p *Provider) GetCacheOn() bool {
	return p.cacheOn
}
