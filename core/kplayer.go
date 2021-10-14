package core

// #include "kplayer.h"
// void goCallBackMessage(char*);
import "C"

import (
    "unsafe"

    kpproto "github.com/bytelang/kplayer/proto"
    "github.com/golang/protobuf/proto"
    log "github.com/sirupsen/logrus"
    "google.golang.org/protobuf/runtime/protoiface"
)

//export goCallBackMessage
// goCallBackMessage define libkplayer callback function
func goCallBackMessage(msgRaw *C.char) {
    msg := C.GoString(msgRaw)
    message := &kpproto.KPMessage{}
    if err := proto.Unmarshal([]byte(msg), message); err != nil {
        log.Fatal("error unmarshal message. error: {}. data: {}", err, msg)
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
    video_width          uint
    video_height         uint
    video_bitrate        uint
    video_qulity         uint
    video_fps            uint
    audio_sample_rate    uint
    audio_channel_layout uint
    audio_channels       uint

    // event message receiver
    callbackFn func(message *kpproto.KPMessage)
}

// GetLibKplayer return singleton LibKplayer instance
func GetLibKplayerInstance() *libKplayer {
    return libKplayerInstance
}

// SetOptions set basic options
func (lb *libKplayer) SetOptions(protocol string, video_width uint, video_height uint, video_bitrate uint, video_qulity uint, video_fps uint, audio_sample_rate uint, audio_channel_layout uint, audio_channels uint) error {
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

func (lb *libKplayer) SendPrompt(action kpproto.EventAction, body protoiface.MessageV1) error {
    str, err := proto.Marshal(body)
    if err != nil {
        return err
    }

    cs := C.CString(string(str))
    defer C.free(unsafe.Pointer(cs))

    C.PromptMessage(C.int(action), cs)
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

    C.ReceiveMessage(C.MessageCallBack(C.goCallBackMessage))

    // start
    stopChan := make(chan bool)
    go func() {
        C.Run()

        log.Info("Core start up success.")
        stopChan <- true
    }()

    <-stopChan
    log.Info("Core shut down success.")
}
