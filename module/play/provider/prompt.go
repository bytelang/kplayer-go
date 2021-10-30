package provider

import (
    "fmt"
    "github.com/bytelang/kplayer/core"
    "github.com/bytelang/kplayer/module"
    "github.com/bytelang/kplayer/types"
    kpproto "github.com/bytelang/kplayer/types/core/proto"
    "github.com/bytelang/kplayer/types/core/proto/msg"
    "github.com/bytelang/kplayer/types/core/proto/prompt"
    svrproto "github.com/bytelang/kplayer/types/server"
)

func (p *Provider) PlayStop(args *svrproto.PlayStopArgs) (*svrproto.PlayStopReply, error) {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_PLAYER_STOP, &prompt.EventPromptPlayerStop{}); err != nil {
        return nil, err
    }

    // register prompt
    endedMsg := &msg.EventMessagePlayerEnded{}
    keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_PLAYER_ENDED, func(msg []byte) bool {
        types.UnmarshalProtoMessage(msg, endedMsg)
        return true
    })
    defer keeperCtx.Close()

    if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
        return nil, err
    }

    // wait context
    keeperCtx.Wait()
    if endedMsg.Error != nil {
        return nil, fmt.Errorf("%s", string(endedMsg.Error))
    }

    return &svrproto.PlayStopReply{}, nil
}
