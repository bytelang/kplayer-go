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

func (p *Provider) OutputAdd(args *svrproto.OutputAddArgs) (*svrproto.OutputAddReply, error) {
	coreKplayer := core.GetLibKplayerInstance()

	if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_OUTPUT_ADD, &prompt.EventPromptOutputAdd{
		Output: &kpproto.PromptOutput{
			Path:   []byte(args.Output.Path),
			Unique: []byte(args.Output.Unique),
		},
	}); err != nil {
		return nil, err
	}

	// register prompt
	outputAddMsg := &msg.EventMessageOutputAdd{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD, func(msg []byte) bool {
		types.UnmarshalProtoMessage(msg, outputAddMsg)
		return string(outputAddMsg.Output.Unique) == args.Output.Unique && string(outputAddMsg.Output.Path) == args.Output.Path
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if outputAddMsg.Error != nil {
		return nil, fmt.Errorf("%s", string(outputAddMsg.Error))
	}

	return &svrproto.OutputAddReply{
		Output: svrproto.Output{
			Path:   string(outputAddMsg.Output.Path),
			Unique: string(outputAddMsg.Output.Unique),
		},
	}, nil
}

func (p *Provider) OutputRemove(args *svrproto.OutputRemoveArgs) (*svrproto.OutputRemoveReply, error) {
	coreKplayer := core.GetLibKplayerInstance()

	if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_OUTPUT_REMOVE, &prompt.EventPromptOutputRemove{
		Unique: []byte(args.Unique),
	}); err != nil {
		return nil, err
	}

	// register prompt
	outputRemoveMsg := &msg.EventMessageOutputRemove{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_OUTPUT_REMOVE, func(msg []byte) bool {
		types.UnmarshalProtoMessage(msg, outputRemoveMsg)
		return string(outputRemoveMsg.Output.Unique) == args.Unique
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if outputRemoveMsg.Error != nil {
		return nil, fmt.Errorf("%s", string(outputRemoveMsg.Error))
	}

	return &svrproto.OutputRemoveReply{
		Output: &svrproto.Output{
			Path:   string(outputRemoveMsg.Output.Path),
			Unique: string(outputRemoveMsg.Output.Unique),
		},
	}, nil
}

func (p *Provider) OutputList(args *svrproto.OutputListArgs) (*svrproto.OutputListReply, error) {
	outputs := []*svrproto.OutputModule{}
	for _, item := range p.outputs {
		outputs = append(outputs, &svrproto.OutputModule{
			Path:       item.Path,
			Unique:     item.Unique,
			CreateTime: item.CreateTime,
			EndTime:    item.EndTime,
			StartTime:  item.StartTime,
			Connected:  item.Connected,
		})
	}

	return &svrproto.OutputListReply{
		Outputs: outputs,
	}, nil
}
