package types

import (
	"encoding/json"
	"fmt"
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/bytelang/kplayer/types/core/proto/msg"
	"github.com/gogo/protobuf/proto"
)

var messageKeyMapping map[kpproto.EventAction]proto.Message

func init() {
	messageKeyMapping = make(map[kpproto.EventAction]proto.Message)

	// register message
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLAYER_STARTED] = &msg.EventMessagePlayerStarted{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLAYER_PAUSE] = &msg.EventMessagePlayerPause{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLAYER_CONTINUE] = &msg.EventMessagePlayerContinue{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLAYER_SKIP] = &msg.EventMessagePlayerSkip{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLAYER_ENDED] = &msg.EventMessagePlayerEnded{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_RESOURCE_START] = &msg.EventMessageResourceStart{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_RESOURCE_FINISH] = &msg.EventMessageResourceFinish{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_RESOURCE_EMPTY] = &msg.EventMessageResourceEmpty{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_RESOURCE_REMOVE] = &msg.EventMessageResourceRemove{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_RESOURCE_ADD] = &msg.EventMessageResourceAdd{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_RESOURCE_LIST] = &msg.EventMessageResourceList{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_RESOURCE_CURRENT] = &msg.EventMessageResourceCurrent{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_OUTPUT_ADD] = &msg.EventMessageOutputAdd{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_OUTPUT_REMOVE] = &msg.EventMessageOutputRemove{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_OUTPUT_LIST] = &msg.EventMessageOutputList{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_OUTPUT_DISCONNECT] = &msg.EventMessageOutputDisconnect{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLUGIN_ADD] = &msg.EventMessagePluginAdd{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLUGIN_REMOVE] = &msg.EventMessagePluginRemove{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLUGIN_LIST] = &msg.EventMessagePluginList{}
	messageKeyMapping[kpproto.EVENT_MESSAGE_ACTION_PLUGIN_UPDATE] = &msg.EventMessagePluginUpdate{}
}

type messageJson struct {
	Action string        `json:"action"`
	Body   proto.Message `json:"body"`
}

func ParseMessageToJson(message kpproto.KPMessage) ([]byte, error) {
	msgCore, ok := messageKeyMapping[message.Action]
	if !ok {
		return nil, fmt.Errorf("message action cannot registed")
	}

	if err := proto.Unmarshal(message.Body, msgCore); err != nil {
		return nil, err
	}

	return json.Marshal(messageJson{
		Action: kpproto.EventAction_name[int32(message.Action)],
		Body:   msgCore,
	})
}
