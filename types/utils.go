package types

import (
    "github.com/golang/protobuf/proto"
    "github.com/google/uuid"
    log "github.com/sirupsen/logrus"
)

func GetRandString(size ...uint) string {
    str := uuid.New().String()
    if len(size) == 0 {
        return str
    }

    return str[:size[0]]
}

func UnmarshalProtoMessage(data []byte, obj proto.Message) {
    if err := proto.Unmarshal(data, obj); err != nil {
        log.Fatal(err)
    }
}
