package provider

import (
	"context"
	"fmt"
	"github.com/bytelang/kplayer/core"
	"github.com/bytelang/kplayer/module"
	kptypes "github.com/bytelang/kplayer/types"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/bytelang/kplayer/types/core/proto/msg"
	kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
	kpmodule "github.com/bytelang/kplayer/types/module"
	svrproto "github.com/bytelang/kplayer/types/server"
	"time"
)

func (p *Provider) OutputAdd(ctx context.Context, args *svrproto.OutputAddArgs) (*svrproto.OutputAddReply, error) {
	outputUnique := args.Output.Unique
	outputPath := args.Output.Path
	if outputUnique == "" {
		outputUnique = kptypes.GetUniqueString(outputPath)
	}

	if err := p.addOutput(kpmodule.Output{
		Path:       outputPath,
		Unique:     outputUnique,
		CreateTime: uint64(time.Now().Unix()),
		Connected:  false,
	}); err != nil {
		return nil, err
	}

	// register prompt
	outputAddMsg := &msg.EventMessageOutputAdd{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_ADD, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, outputAddMsg)
		re := outputAddMsg.Output.Unique == outputUnique && outputAddMsg.Output.Path == outputPath
		return re
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
		Output: &svrproto.Output{
			Path:   outputAddMsg.Output.Path,
			Unique: outputAddMsg.Output.Unique,
		},
	}, nil
}

func (p *Provider) OutputRemove(ctx context.Context, args *svrproto.OutputRemoveArgs) (*svrproto.OutputRemoveReply, error) {
	if !p.configList.Exist(args.Unique) {
		return nil, OutputUniqueNotFound
	}
	output, _, err := p.configList.GetOutputByUnique(args.Unique)
	if err != nil {
		return nil, err
	}
	if output.Connected == false {
		removeOutput, err := p.configList.RemoveOutputByUnique(output.Unique)
		if err != nil {
			return nil, err
		}
		return &svrproto.OutputRemoveReply{
			Output: &svrproto.Output{
				Path:   removeOutput.Path,
				Unique: removeOutput.Unique,
			},
		}, nil
	}

	coreKplayer := core.GetLibKplayerInstance()

	if err := coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_OUTPUT_REMOVE, &kpprompt.EventPromptOutputRemove{
		Unique: args.Unique,
	}); err != nil {
		return nil, err
	}

	// register prompt
	outputRemoveMsg := &msg.EventMessageOutputRemove{}
	keeperCtx := module.NewKeeperContext(kptypes.GetRandString(), kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_REMOVE, func(msg string) bool {
		kptypes.UnmarshalProtoMessage(msg, outputRemoveMsg)
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

func (p *Provider) OutputList(ctx context.Context, args *svrproto.OutputListArgs) (*svrproto.OutputListReply, error) {
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
