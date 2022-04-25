package core

// #cgo LDFLAGS: -lkplayer -lkpcodec -lkputil -lkpadapter -lkpplugin
// #include "extra.h"
// void goCallBackMessage(int, char*);
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
	libKplayerInstance.callbackFn(ac, msg)
}

var libKplayerInstance *libKplayer = &libKplayer{
	callbackFn: func(action int, message string) {},
}

// libKplayer
type libKplayer struct {
	// basic params
	protocol             string
	video_width          uint32
	video_height         uint32
	video_bitrate        uint32
	video_qulity         uint32
	video_fps            uint32
	audio_sample_rate    uint32
	audio_channel_layout uint32
	audio_channels       uint32

	// options
	cache_on              bool
	skip_invalid_resource bool

	// event message receiver
	callbackFn func(action int, message string)

	// delay queue size
	delay_queue_size uint16

	// fill strategy
	fill_strategy int32
}

// GetLibKplayer return singleton LibKplayer instance
func GetLibKplayerInstance() *libKplayer {
	return libKplayerInstance
}

// SetOptions set basic options
func (lb *libKplayer) SetOptions(protocol string,
	video_width uint32,
	video_height uint32,
	video_bitrate uint32,
	video_qulity uint32,
	video_fps uint32,
	audio_sample_rate uint32,
	audio_channel_layout uint32,
	audio_channels uint32, delay_queue_size uint32, fill_strategy int32) error {
	libKplayerInstance.protocol = strings.ToLower(protocol)
	libKplayerInstance.video_width = video_width
	libKplayerInstance.video_height = video_height
	libKplayerInstance.video_bitrate = video_bitrate
	libKplayerInstance.video_qulity = video_qulity
	libKplayerInstance.video_fps = video_fps
	libKplayerInstance.audio_sample_rate = audio_sample_rate
	libKplayerInstance.audio_channel_layout = audio_channel_layout
	libKplayerInstance.audio_channels = audio_channels

	// other params
	libKplayerInstance.delay_queue_size = uint16(delay_queue_size)
	libKplayerInstance.fill_strategy = fill_strategy

	return nil
}

func (lb *libKplayer) SetCallBackMessage(fn func(action int, message string)) {
	lb.callbackFn = fn
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

	C.Initialization(C.CString(lb.protocol),
		C.int(lb.video_width),
		C.int(lb.video_height),
		C.int(lb.video_bitrate),
		C.int(lb.video_qulity),
		C.int(lb.video_fps),
		C.int(lb.audio_sample_rate),
		C.int(lb.audio_channel_layout),
		C.int(lb.audio_channels),
		C.short(lb.delay_queue_size),
		C.int(lb.fill_strategy))
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
