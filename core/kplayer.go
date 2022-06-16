package core

// #cgo LDFLAGS: -lkplayer -lkpcodec -lkputil -lkpadapter -lkpplugin
// #include "extra.h"
// void goCallBackMessage(int, char*);
// void goCallBackProgress(double, int);
import "C"

import (
	"bytes"
	"fmt"
	kpprompt "github.com/bytelang/kplayer/types/core/proto/prompt"
	"github.com/golang/protobuf/jsonpb"
	"strings"
	"unsafe"

	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

//export goCallBackMessage
// goCallBackMessage define libkplayer callback function
func goCallBackMessage(action C.int, msgRaw *C.char) {
	msg := C.GoString(msgRaw)
	ac := int(action)
	libKplayerInstance.callbackMessageFn(ac, msg)
}

//export goCallBackProgress
// goCallBackProgress define libkplayer callback function
func goCallBackProgress(percent C.double, bitRate C.int) {
	libKplayerInstance.callbackProgressFn(float64(percent), int(bitRate))
}

type CoreKplayerOption string

var (
	ProtocolOption     CoreKplayerOption = "protocol"
	VideoWidthOption   CoreKplayerOption = "video_width"
	VideoHeightOption  CoreKplayerOption = "video_height"
	VideoBitrateOption CoreKplayerOption = "video_bitrate"
	VideoQualityOption CoreKplayerOption = "video_quality"
	VideoFpsOption     CoreKplayerOption = "video_fps"
	AudioSampleRate    CoreKplayerOption = "audio_sample_rate"
	AudioChannelLayout CoreKplayerOption = "audio_channel_layout"
	AudioChannels      CoreKplayerOption = "audio_channels"
	VideoFillStrategy  CoreKplayerOption = "video_fill_strategy"
)

var libKplayerInstance *libKplayer = &libKplayer{
	protocol:              "file",
	video_width:           848,
	video_height:          480,
	video_bitrate:         0,
	video_quality:         0,
	video_fps:             25,
	audio_sample_rate:     44100,
	audio_channel_layout:  3,
	audio_channels:        2,
	cache_on:              false,
	skip_invalid_resource: false,
	video_fill_strategy:   0,
	callbackMessageFn:     func(action int, message string) {},
	callbackProgressFn:    func(percent float64, bitRate int) {},
}

// libKplayer
type libKplayer struct {
	// basic params
	protocol             string
	video_width          uint32
	video_height         uint32
	video_bitrate        uint32
	video_quality        uint32
	video_fps            uint32
	audio_sample_rate    uint32
	audio_channel_layout uint32
	audio_channels       uint32
	video_fill_strategy  int32

	// options
	cache_on              bool
	skip_invalid_resource bool

	// event message receiver
	callbackMessageFn  func(action int, message string)
	callbackProgressFn func(percent float64, bitRate int)
}

// GetLibKplayer return singleton LibKplayer instance
func GetLibKplayerInstance() *libKplayer {
	return libKplayerInstance
}

// SetOptions set basic options
func (lb *libKplayer) SetOptions(options map[CoreKplayerOption]interface{}) error {
	for option, value := range options {
		switch option {
		case ProtocolOption:
			libKplayerInstance.protocol = strings.ToLower(value.(string))
		case VideoWidthOption:
			libKplayerInstance.video_width = value.(uint32)
		case VideoHeightOption:
			libKplayerInstance.video_height = value.(uint32)
		case VideoBitrateOption:
			libKplayerInstance.video_bitrate = value.(uint32)
		case VideoQualityOption:
			libKplayerInstance.video_quality = value.(uint32)
		case VideoFpsOption:
			libKplayerInstance.video_fps = value.(uint32)
		case VideoFillStrategy:
			libKplayerInstance.video_fill_strategy = value.(int32)
		case AudioSampleRate:
			libKplayerInstance.audio_sample_rate = value.(uint32)
		case AudioChannelLayout:
			libKplayerInstance.audio_channel_layout = value.(uint32)
		case AudioChannels:
			libKplayerInstance.audio_channels = value.(uint32)
		}
	}
	return nil
}

func (lb *libKplayer) SetCallBackMessage(fn func(action int, message string)) {
	lb.callbackMessageFn = fn
}

func (lb *libKplayer) SetCallBackProgress(fn func(percent float64, bitRate int)) {
	lb.callbackProgressFn = fn
}

func (lb *libKplayer) GetInformation() *kpproto.Information {
	infoMemorySize := 2000
	str := make([]byte, infoMemorySize)
	cs := (*C.char)(unsafe.Pointer(&str[0]))

	C.GetInformation(cs, C.int(infoMemorySize))

	str = bytes.Trim(str, "\x00")
	info := &kpproto.Information{}
	if err := jsonpb.UnmarshalString(string(str), info); err != nil {
		log.Fatalf("error: %s", err)
	}
	return info
}

func (lb *libKplayer) SendPrompt(action kpproto.EventPromptAction, body proto.Message) error {
	m := jsonpb.Marshaler{}
	str, err := m.MarshalToString(body)
	if err != nil {
		return err
	}

	cs := C.CString(str)
	defer C.free(unsafe.Pointer(cs))

	C.PromptMessage(C.int(action), cs)
	log.WithFields(log.Fields{"action": kpproto.EventPromptAction_name[int32(action)]}).Debug("send prompt message")
	return nil
}

func (lb *libKplayer) Run() {
	// start
	stopChan := make(chan bool)
	go func() {
		defer func() {
			stopChan <- true
		}()

		log.Info("core start up success")
		result := C.Run()

		if int(result) < 0 {
			log.WithFields(log.Fields{"code": int(result), "error": C.GoString(C.GetError())}).Error("core return error")
		}
	}()

	<-stopChan
	log.Info("core shut down success")
}

func (lb *libKplayer) SetCacheOn(c bool) {
	lb.cache_on = c
}

func (lb *libKplayer) SetSkipInvalidResource(s bool) {
	lb.skip_invalid_resource = s
}

func (lb *libKplayer) SetLogLevel(path string, level int) {
	logPath := C.CString(path)
	C.SetLogLevel(logPath, C.int(level))
	defer C.free(unsafe.Pointer(logPath))
}

func (lb *libKplayer) Initialization() {
	if lb.cache_on {
		C.SetCacheOn(C.int(1))
	}
	if lb.skip_invalid_resource {
		C.SetSkipInvalidResource(C.int(1))
	}

	C.ReceiveMessage(C.MessageCallBack(C.goCallBackMessage))
	C.ProgressCallback(C.ProgressCallBack(C.goCallBackProgress))

	C.Initialization(C.CString(lb.protocol),
		C.int(lb.video_width),
		C.int(lb.video_height),
		C.int(lb.video_bitrate),
		C.int(lb.video_quality),
		C.int(lb.video_fps),
		C.int(lb.audio_sample_rate),
		C.int(lb.audio_channel_layout),
		C.int(lb.audio_channels),
		C.int(lb.video_fill_strategy))
}

func (lb *libKplayer) AddOutput(body *kpprompt.EventPromptOutputAdd) error {
	m := jsonpb.Marshaler{}
	str, err := m.MarshalToString(body)
	if err != nil {
		return err
	}

	cs := C.CString(str)
	defer C.free(unsafe.Pointer(cs))
	resultCode := C.AddOutput(cs)
	if resultCode != 0 {
		return fmt.Errorf("add output failed. result code: %d", resultCode)
	}

	return nil
}

func (lb *libKplayer) AddPlugin(body *kpprompt.EventPromptPluginAdd) error {
	m := jsonpb.Marshaler{}
	str, err := m.MarshalToString(body)
	if err != nil {
		return err
	}

	cs := C.CString(str)
	defer C.free(unsafe.Pointer(cs))
	resultCode := C.AddPlugin(cs)
	if resultCode != 0 {
		return fmt.Errorf("add plugin failed. result code: %d", resultCode)
	}

	return nil
}
