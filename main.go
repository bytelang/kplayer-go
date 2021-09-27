package main

import (
    "github.com/bytelang/kplayer/core"
    kpproto "github.com/bytelang/kplayer/proto"
    "github.com/bytelang/kplayer/proto/prompt"
    log "github.com/sirupsen/logrus"
    "os"
)

func init() {
    log.SetOutput(os.Stdout)
    log.SetReportCaller(true)
    log.SetLevel(log.TraceLevel)
}

func main() {
    coreKplayer := core.GetLibKplayerInstance()
    if err := coreKplayer.SetOptions("rtmp", 800, 480, 0, 0, 30, 48000, 3, 2); err != nil {
        log.Fatal(err)
    }

    coreKplayer.SetCallBackMessage(MessageConsumer)
    coreKplayer.Run()
}

func MessageConsumer(message *kpproto.KPMessage) {
    log.Debug("receive broadcast message: ", message.Action)

    // global core
    coreKplayer := core.GetLibKplayerInstance()

    var err error
    switch message.Action {
    case kpproto.EventAction_EVENT_MESSAGE_ACTION_PLAYER_STARTED:
        err = coreKplayer.SendPrompt(kpproto.EventAction_EVENT_PROMPT_ACTION_OUTPUT_ADD, &prompt.EventPromptOutputAdd{
            Path:   "output.flv",
            Unique: "test",
        })
    }

    if err != nil {
        log.Errorf("send prompt command failed. error: %s", err)
    }
}
