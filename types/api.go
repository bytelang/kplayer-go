package types

import (
	"encoding/json"
	"fmt"
	"github.com/bytelang/kplayer/types/api"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	ApiScheme  string = "https"
	ApiHost    string = ""
	ApiPort    string = ""
	ApiVersion string = ""
)

type ApiError string

func (ae ApiError) Error() string {
	return string(ae)
}

func GetTlsHttpClient() *http.Client {
	config, err := GetTlsClientConfig()
	if err != nil {
		panic(err)
	}

	transPort := &http.Transport{
		TLSClientConfig: config,
	}

	return &http.Client{Transport: transPort, Timeout: time.Second * 10}
}

func GetApiRequestUrl(path string) string {
	return fmt.Sprintf("%s://%s:%s/%s%s", ApiScheme, ApiHost, ApiPort, ApiVersion, path)
}

func RequestHttpGet(host string, params proto.Message, message proto.Message) error {
	d, err := json.Marshal(params)
	if err != nil {
		return err
	}

	mapping := make(map[string]string)
	if err := json.Unmarshal(d, &mapping); err != nil {
		return err
	}

	var query string
	for key, item := range mapping {
		query = query + fmt.Sprintf("%s=%s&", url.QueryEscape(key), url.QueryEscape(item))
	}

	// request
	resp, err := GetTlsHttpClient().Get(fmt.Sprintf("%s?%s", host, query))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("response status code: %d", resp.StatusCode)
		}
		return ApiError(fmt.Sprintf("response status code: %d", resp.StatusCode))
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// unmarshal
	if err := json.Unmarshal(respBody, message); err != nil {
		return err
	}

	return nil
}

// plugin
func GetPlugin(request *api.PluginInformationRequest) (*api.PluginInformationResponse, error) {
	resp := &api.PluginInformationResponse{}
	if err := RequestHttpGet(GetApiRequestUrl("/plugin/info"), request, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// resource
func GetResource(request *api.ResourceInformationRequest) (*api.ResourceInformationResponse, error) {
	resp := &api.ResourceInformationResponse{}
	if err := RequestHttpGet(GetApiRequestUrl("/resource/info"), request, resp); err != nil {
		return nil, err
	}

	return resp, nil
}

// knock
func Knock() error {
	return RequestHttpGet(GetApiRequestUrl("/status/knock"), &api.StatusKnockRequest{}, &api.StatusKnockResponse{})
}
