package types

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/runtime/protoiface"
	"io"
	"net/http"
	"os"
)

var issueRandStr map[string]bool

func init() {
	issueRandStr = make(map[string]bool)
}

func GetRandString(size ...uint) string {
	var str string
	for {
		str = uuid.New().String()
		if len(size) != 0 {
			str = str[:size[0]]
		}
		if ok := issueRandStr[str]; !ok {
			break
		}
	}

	return str
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

func MkDir(dir string) error {
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return os.Mkdir(dir, os.ModePerm)
	}
	if stat.IsDir() {
		return nil
	}

	return fmt.Errorf("plugin directory can not be avaiable")
}

func DownloadFile(url, filePath string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	openFile, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer openFile.Close()

	if _, err := io.Copy(openFile, res.Body); err != nil {
		return err
	}

	return nil
}
