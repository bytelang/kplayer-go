package core

// #cgo LDFLAGS: -lkplayer -lkpcodec -lkputil -lkpadapter -lkpplugin
// #include "extra.h"
// void goCallBackMessage(char*);
import "C"

import (
	"bytes"
	"unsafe"

	kpproto "github.com/bytelang/kplayer/types/core/proto"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

//export goCallBackMessage
// goCallBackMessage define libkplayer callback function
func goCallBackMessage(msgRaw *C.char) {
	msg := C.GoString(msgRaw)
	message := &kpproto.KPMessage{}
	if err := proto.Unmarshal([]byte(msg), message); err != nil {
		log.WithFields(log.Fields{"error": err, "message": msg}).Fatal("error unmarshal message")
	}

	libKplayerInstance.callbackFn(message)
}

var libKplayerInstance *libKplayer = &libKplayer{
	callbackFn: func(message *kpproto.KPMessage) {},
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
	callbackFn func(message *kpproto.KPMessage)
}

// GetLibKplayer return singleton LibKplayer instance
func GetLibKplayerInstance() *libKplayer {
	return libKplayerInstance
}

// SetOptions set basic options
func (lb *libKplayer) SetOptions(protocol string, video_width uint32, video_height uint32, video_bitrate uint32, video_qulity uint32, video_fps uint32, audio_sample_rate uint32, audio_channel_layout uint32, audio_channels uint32) error {
	libKplayerInstance.protocol = protocol
	libKplayerInstance.video_width = video_width
	libKplayerInstance.video_height = video_height
	libKplayerInstance.video_bitrate = video_bitrate
	libKplayerInstance.video_qulity = video_qulity
	libKplayerInstance.video_fps = video_fps
	libKplayerInstance.audio_sample_rate = audio_sample_rate
	libKplayerInstance.audio_channel_layout = audio_channel_layout
	libKplayerInstance.audio_channels = audio_channels

	return nil
}

func (lb *libKplayer) SetCallBackMessage(fn func(message *kpproto.KPMessage)) {
	lb.callbackFn = fn
}

func (lb *libKplayer) GetInformation() *kpproto.Information {
	infoMemorySize := 200
	str := make([]byte, infoMemorySize)
	cs := (*C.char)(unsafe.Pointer(&str[0]))

	C.GetInformation(cs, C.int(infoMemorySize))

	str = bytes.Trim(str, "\x00")
	info := &kpproto.Information{}
	if err := proto.Unmarshal(str, info); err != nil {
		log.Fatalf("error: %s", err)
	}
	return info
}

func (lb *libKplayer) SendPrompt(action kpproto.EventAction, body proto.Message) error {
	str, err := proto.Marshal(body)
	if err != nil {
		return err
	}

	cs := C.CString(string(str))
	defer C.free(unsafe.Pointer(cs))

	C.PromptMessage(C.int(action), cs)
	log.WithFields(log.Fields{"action": kpproto.EventAction_name[int32(action)]}).Debug("send prompt message")
	return nil
}

func (lb *libKplayer) Run() {
	C.Initialization(C.CString(lb.protocol),
		C.int(lb.video_width),
		C.int(lb.video_height),
		C.int(lb.video_bitrate),
		C.int(lb.video_qulity),
		C.int(lb.video_fps),
		C.int(lb.audio_sample_rate),
		C.int(lb.audio_channel_layout),
		C.int(lb.audio_channels))

	if lb.cache_on {
		C.SetCacheOn(C.int(1))
	}
	if lb.skip_invalid_resource {
		C.SetSkipInvalidResource(C.int(1))
	}

	C.ReceiveMessage(C.MessageCallBack(C.goCallBackMessage))

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