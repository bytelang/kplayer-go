package types

import (
    "github.com/golang/protobuf/proto"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
    "google.golang.org/protobuf/runtime/protoiface"
)

func GetRandString(size ...uint) string {
    str := uuid.New().String()
    if len(size) == 0 {
        return str
    }

    return str[:size[0]]
}

func UnmarshalProtoMessage(data []byte, obj protoiface.MessageV1) {
    if err := proto.Unmarshal(data, obj); err != nil {
        log.Fatal("error unmarshal message. error: %s. data: %s", err, string(data))
    }
}
