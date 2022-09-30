package core

import (
	kpproto "github.com/bytelang/kplayer/types/core/proto"
	kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
	"testing"
)

func TestFilePlay(t *testing.T) {
	coreKplayer := GetLibKplayerInstance()
	coreKplayer.Initialization()

	coreKplayer.SetCallBackMessage(func(action int, message string) {
		switch kpproto.EventMessageAction(action) {
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY:
			//add output
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_OUTPUT_ADD, &kpprompt.EventPromptOutputAdd{Output: &kpprompt.PromptOutput{
				Path:   "test.flv",
				Unique: "test",
			}})

			// add resource
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{Resource: &kpproto.PromptResource{
				Path:   "short.flv",
				Unique: "test",
			}})
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &kpprompt.EventPromptPlayerStop{})
		}
	})

	result := coreKplayer.Run()
	if result != 0 {
		t.Fatal("running core kplayer failed")
	}
}

func TestRtmpPlay(t *testing.T) {
	coreKplayer := GetLibKplayerInstance()

	coreKplayer.SetOptions(map[CoreKplayerOption]interface{}{
		ProtocolOption: "RTMP",
	})

	coreKplayer.Initialization()
	coreKplayer.SetCallBackMessage(func(action int, message string) {
		switch kpproto.EventMessageAction(action) {
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY:
			//add output
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_OUTPUT_ADD, &kpprompt.EventPromptOutputAdd{Output: &kpprompt.PromptOutput{
				Path:   "rtmp://127.0.0.1:1935/live/test",
				Unique: "test",
			}})

			// add resource
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{Resource: &kpproto.PromptResource{
				Path:   "short.flv",
				Unique: "test",
			}})
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &kpprompt.EventPromptPlayerStop{})
		}
	})

	result := coreKplayer.Run()
	if result != 0 {
		t.Fatal("running core kplayer failed")
	}
}

func TestMultiFilePlay(t *testing.T) {
	coreKplayer := GetLibKplayerInstance()

	coreKplayer.Initialization()

	end := false

	coreKplayer.SetCallBackMessage(func(action int, message string) {
		switch kpproto.EventMessageAction(action) {
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY:
			//add output
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_OUTPUT_ADD, &kpprompt.EventPromptOutputAdd{Output: &kpprompt.PromptOutput{
				Path:   "test.flv",
				Unique: "test",
			}})

			// add resource
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{Resource: &kpproto.PromptResource{
				Path:   "short.flv",
				Unique: "test",
			}})
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
			if end {
				coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &kpprompt.EventPromptPlayerStop{})
				break
			}

			// add next resource
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{Resource: &kpproto.PromptResource{
				Path:   "short.flv",
				Unique: "test",
			}})
			end = true
		}
	})

	result := coreKplayer.Run()
	if result != 0 {
		t.Fatal("running core kplayer failed")
	}
}

func TestLibKplayer_SetSkipInvalidResource(t *testing.T) {
	coreKplayer := GetLibKplayerInstance()

	coreKplayer.SetSkipInvalidResource(true)
	coreKplayer.Initialization()
	end := false

	coreKplayer.SetCallBackMessage(func(action int, message string) {
		switch kpproto.EventMessageAction(action) {
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY:
			// add resource
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{Resource: &kpproto.PromptResource{
				Path:   "invalid.flv",
				Unique: "test",
			}})
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &kpprompt.EventPromptPlayerStop{})
			end = true
		}
	})

	result := coreKplayer.Run()
	if result != 0 {
		t.Fatal("running core kplayer failed")
	}

	if !end {
		t.Fatal("invalid resource failed")
	}
}

func TestLibKplayer_GenerateCache(t *testing.T) {
	// generate cache
	coreKplayer := GetLibKplayerInstance()
	coreKplayer.SetCacheOn(true)
	coreKplayer.Initialization()
	end := false

	coreKplayer.SetCallBackMessage(func(action int, message string) {
		switch kpproto.EventMessageAction(action) {
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY:
			// add resource
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{Resource: &kpproto.PromptResource{
				Path:   "test.flv",
				Unique: "test",
			}})
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &kpprompt.EventPromptPlayerStop{})
			end = true
		}
	})

	result := coreKplayer.Run()
	if result != 0 {
		t.Fatal("running core kplayer failed")
	}
	if !end {
		t.Fatal("invalid resource failed")
	}
}

func TestLibKplayer_SetCacheUncheckSource(t *testing.T) {
	coreKplayer := GetLibKplayerInstance()

	coreKplayer.SetCacheUncheckSource()
	coreKplayer.SetCacheOn(true)
	coreKplayer.Initialization()
	end := false

	coreKplayer.SetCallBackMessage(func(action int, message string) {
		switch kpproto.EventMessageAction(action) {
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY:
			// add resource
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{Resource: &kpproto.PromptResource{
				Path:   "/invalid/test.flv",
				Unique: "test",
			}})
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &kpprompt.EventPromptPlayerStop{})
			end = true
		}
	})

	result := coreKplayer.Run()
	if result != 0 {
		t.Fatal("running core kplayer failed")
	}

	if !end {
		t.Fatal("invalid resource failed")
	}
}

func TestLibKplayer_InvalidResource(t *testing.T) {
	coreKplayer := GetLibKplayerInstance()

	coreKplayer.SetCacheUncheckSource()
	coreKplayer.SetCacheOn(true)
	coreKplayer.Initialization()

	coreKplayer.SetCallBackMessage(func(action int, message string) {
		switch kpproto.EventMessageAction(action) {
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_EMPTY:
			// add resource
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_RESOURCE_ADD, &kpprompt.EventPromptResourceAdd{Resource: &kpproto.PromptResource{
				Path:   "invalid.flv",
				Unique: "test",
			}})
		case kpproto.EventMessageAction_EVENT_MESSAGE_ACTION_RESOURCE_FINISH:
			coreKplayer.SendPrompt(kpproto.EventPromptAction_EVENT_PROMPT_ACTION_PLAYER_STOP, &kpprompt.EventPromptPlayerStop{})
		}
	})

	result := coreKplayer.Run()
	if result == 0 {
		t.Fatal("running core kplayer failed")
	}
}
