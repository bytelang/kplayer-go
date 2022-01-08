package types

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
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

func UnmarshalProtoMessage(data string, obj protoiface.MessageV1) {
	if err := jsonpb.UnmarshalString(data, obj); err != nil {
		log.WithFields(log.Fields{"error": err, "data": data}).Fatal("error unmarshal message")
	}
}

func MarshalProtoMessage(obj proto.Message) (string, error) {
	m := jsonpb.Marshaler{}
	d, err := m.MarshalToString(obj)
	if err != nil {
		return "", err
	}

	return d, nil
}

func CopyProtoMessage(src protoiface.MessageV1, dst protoiface.MessageV1) error {
	d, err := MarshalProtoMessage(src)
	if err != nil {
		return err
	}

	UnmarshalProtoMessage(d, dst)
	return nil
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
