package client

import (
	"encoding/json"
	"fmt"
	"github.com/bytelang/kplayer/types/config"
	"github.com/gogo/protobuf/proto"
	"io/ioutil"
	"net/http"
	"strings"
)

type body struct {
	JsonRpc string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Id      string          `json:"id"`
	Params  []proto.Message `json:"params"`
}

type result struct {
	Result proto.Message `json:"result"`
	Error  *string       `json:"error"`
	Id     string        `json:"id"`
}

func ClientRequest(rpc *config.Rpc, method string, request proto.Message, response proto.Message) error {
	if !rpc.On {
		return fmt.Errorf("rpc server not start up")
	}

	bodyStruct := body{
		JsonRpc: "2.0",
		Method:  method,
		Id:      "1",
		Params:  []proto.Message{request},
	}
	bodyContent, err := json.Marshal(bodyStruct)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://127.0.0.1:%d/rpc", rpc.Port), strings.NewReader(string(bodyContent)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	// request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	res := &result{
		Result: response,
	}
	if err := json.Unmarshal(body, res); err != nil {
		return err
	}

	if res.Error != nil {
		return fmt.Errorf("%s", *res.Error)
	}

	return nil
}
