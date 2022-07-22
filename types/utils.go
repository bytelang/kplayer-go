package types

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/ghodss/yaml"
	"github.com/go-playground/validator/v10"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/runtime/protoiface"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var issueRandStr map[string]bool

const (
	VAL   = 0x3FFFFFFF
	INDEX = 0x0000003D
)

var (
	alphabet = []byte("abcdefghijklmnopqrstuvwxyz0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ")
)

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

func GetUniqueString(str string) string {
	uniqueStr := ShortNameGenerate(str)[0]
	if ok := issueRandStr[uniqueStr]; !ok {
		issueRandStr[uniqueStr] = true
		return uniqueStr
	}

	basicUniqueStr := uniqueStr
	for i := 1; i <= 100; i++ {
		uniqueStr = fmt.Sprintf("%s-%d", basicUniqueStr, i)
		if ok := issueRandStr[uniqueStr]; !ok {
			issueRandStr[uniqueStr] = true
			return uniqueStr
		}
	}

	log.Fatal("generate unique string failed")
	return ""
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

func GetDirectorFiles(dir string) ([]string, error) {
	stat, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("not directory")
	}

	filesInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	files := []string{}
	for _, item := range filesInfo {
		if !item.IsDir() {
			files = append(files, filepath.Join(dir, item.Name()))
		}
	}

	return files, nil
}

func DownloadFile(url, filePath string) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}
	if err := MkDir(filepath.Dir(filePath)); err != nil {
		log.WithFields(log.Fields{"file path": filePath}).Error(err)
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

func ReadPlugin(filePath string) ([]byte, error) {
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	if string(fileContent[:len(KplayerPluginSignHeader)]) == KplayerPluginSignHeader {
		encryptData := fileContent[len(KplayerPluginSignHeader):]

		// aes decrypt
		decryptData, err := openssl.AesCBCDecrypt(encryptData, []byte(CipherKey), []byte(CipherIV), openssl.PKCS5_PADDING)
		if err != nil {
			log.Fatal(err)
		}
		return decryptData, nil
	}

	return fileContent, nil
}

func FormatYamlProtoMessage(msg proto.Message) (string, error) {
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return "", err
	}

	yamlData, err := yaml.JSONToYAML(jsonData)
	if err != nil {
		return "", err
	}

	return string(yamlData), nil
}

func GetClientContextFromCommand(cmd *cobra.Command) *ClientContext {
	var clientCtx *ClientContext
	if ptr, err := GetCommandContext(cmd, ClientContextKey); err != nil {
		log.Fatalf("get client context failed. error: %s", err)
	} else {
		clientCtx = ptr.(*ClientContext)
	}

	return clientCtx
}

func ArrayInString(arr []string, search string) bool {
	for _, item := range arr {
		if item == search {
			return true
		}
	}

	return false
}

func TrimCRLF(replaceStr string) string {
	replaceStr = strings.ReplaceAll(replaceStr, "\n", "\\n")
	replaceStr = strings.ReplaceAll(replaceStr, "\r", "")

	return replaceStr
}

func ShortNameGenerate(longURL string) [4]string {
	md5Str := getMd5Str(longURL)
	//var hexVal int64
	var tempVal int64
	var result [4]string
	var tempUri []byte
	for i := 0; i < 4; i++ {
		tempSubStr := md5Str[i*8 : (i+1)*8]
		hexVal, err := strconv.ParseInt(tempSubStr, 16, 64)
		if err != nil {
			return result
		}
		tempVal = int64(VAL) & hexVal
		var index int64
		tempUri = []byte{}
		for i := 0; i < 6; i++ {
			index = INDEX & tempVal
			tempUri = append(tempUri, alphabet[index])
			tempVal = tempVal >> 5
		}
		result[i] = string(tempUri)
	}
	return result
}

func ValidateStructor(in interface{}) error {
	validate := validator.New()
	if err := validate.Struct(in); err != nil {
		return err
	}
	return nil
}

func getMd5Str(str string) string {
	m := md5.New()
	m.Write([]byte(str))
	c := m.Sum(nil)
	return hex.EncodeToString(c)
}
