package types

import (
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/runtime/protoiface"
	"os"
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
		log.WithFields(log.Fields{"error": err, "data": string(data)}).Fatal("error unmarshal message")
	}
}

func CopyProtoMessage(src protoiface.MessageV1, dst protoiface.MessageV1) error {
	d, err := proto.Marshal(src)
	if err != nil {
		return err
	}

	return proto.Unmarshal(d, dst)
}

type KPString struct {
	d []byte
}

func NewKPString(d []byte) *KPString {
	return &KPString{d: d}
}

func (ks *KPString) Equal(con string) bool {
	return string(ks.d) == con
}

func (ks *KPString) String() string {
	return string(ks.d)
}

func FileExists(filePath string) bool {
	stat, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	if stat.IsDir() {
		return false
	}

	return true
}
