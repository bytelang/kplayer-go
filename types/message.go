package types

import (
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/bytelang/kplayer/types/core/proto/msg"
	"github.com/gogo/protobuf/proto"
)

var messageKeyMapping map[kpproto.EventMessageAction]proto.Message

func init() {
	messageKeyMapping = make(map[kpproto.EventMessageAction]proto.Message)

	// register message
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_STARTED] = &msg.EventMessagePlayerStarted{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_PAUSE] = &msg.EventMessagePlayerPause{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_CONTINUE] = &msg.EventMessagePlayerContinue{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_SKIP] = &msg.EventMessagePlayerSkip{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLAYER_ENDED] = &msg.EventMessagePlayerEnded{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_START] = &msg.EventMessageResourceStart{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH] = &msg.EventMessageResourceFinish{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY] = &msg.EventMessageResourceEmpty{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_REMOVE] = &msg.EventMessageResourceRemove{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_ADD] = &msg.EventMessageResourceAdd{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_LIST] = &msg.EventMessageResourceList{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_CURRENT] = &msg.EventMessageResourceCurrent{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_ADD] = &msg.EventMessageOutputAdd{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_REMOVE] = &msg.EventMessageOutputRemove{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_LIST] = &msg.EventMessageOutputList{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_OUTPUT_DISCONNECT] = &msg.EventMessageOutputDisconnect{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_ADD] = &msg.EventMessagePluginAdd{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_REMOVE] = &msg.EventMessagePluginRemove{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_LIST] = &msg.EventMessagePluginList{}
	messageKeyMapping[kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_PLUGIN_UPDATE] = &msg.EventMessagePluginUpdate{}
}

type messageJson struct {
	Action string        `json:"action"`
	Body   proto.Message `json:"body"`
}

func ParseMessageToJson(message proto.Message) (string, error) {
	return MarshalProtoMessage(message)
}
