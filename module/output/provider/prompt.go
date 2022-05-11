package provider

import (
	"fmt"
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	"github.com/bytelang/kplayer/types"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/bytelang/kplayer/types/core/proto/msg"
	kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
	kpmodule "github.com/bytelang/kplayer/types/module"
	svrproto "github.com/bytelang/kplayer/types/server"
	"time"
)

func (p *Provider) OutputAdd(args *svrproto.OutputAddArgs) (*svrproto.OutputAddReply, error) {
	if err := p.addOutput(kpmodule.Output{
		Path:       args.Output.Path,
		Unique:     args.Output.Unique,
		CreateTime: uint64(time.Now().Unix()),
		Connected:  false,
	}); err != nil {
		return nil, err
	}

	// register prompt
	outputAddMsg := &msg.EventMessageOutputAdd{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD, func(msg string) bool {
		types.UnmarshalProtoMessage(msg, outputAddMsg)
		return outputAddMsg.Output.Unique == args.Output.Unique && outputAddMsg.Output.Path == args.Output.Path
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(outputAddMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", outputAddMsg.Error)
	}

	return &svrproto.OutputAddReply{
		Output: svrproto.Output{
			Path:   outputAddMsg.Output.Path,
			Unique: outputAddMsg.Output.Unique,
		},
	}, nil
}

func (p *Provider) OutputRemove(args *svrproto.OutputRemoveArgs) (*svrproto.OutputRemoveReply, error) {
	coreKplayer := core.GetLibKplayerInstance()

	if err := coreKplayer.SendPrompt(kpproto.EVENT_PROMPT_ACTION_OUTPUT_REMOVE, &kpprompt.EventPromptOutputRemove{
		Unique: args.Unique,
	}); err != nil {
		return nil, err
	}

	// register prompt
	outputRemoveMsg := &msg.EventMessageOutputRemove{}
	keeperCtx := module.NewKeeperContext(types.GetRandString(), kpproto.EVENT_MESSAGE_ACTION_OUTPUT_REMOVE, func(msg string) bool {
		types.UnmarshalProtoMessage(msg, outputRemoveMsg)
		return outputRemoveMsg.Output.Unique == args.Unique
	})
	defer keeperCtx.Close()

	if err := p.RegisterKeeperChannel(keeperCtx); err != nil {
		return nil, err
	}

	// wait context
	keeperCtx.Wait()
	if len(outputRemoveMsg.Error) != 0 {
		return nil, fmt.Errorf("%s", outputRemoveMsg.Error)
	}

	return &svrproto.OutputRemoveReply{
		Output: &svrproto.Output{
			Path:   outputRemoveMsg.Output.Path,
			Unique: outputRemoveMsg.Output.Unique,
		},
	}, nil
}

func (p *Provider) OutputList(args *svrproto.OutputListArgs) (*svrproto.OutputListReply, error) {
	outputs := []*svrproto.OutputModule{}
	for _, item := range p.configList.outputs {
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
