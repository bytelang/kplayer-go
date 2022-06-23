package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

// output
func TestPluginList(t *testing.T) {
	resp, err := getClient().Get(Host + "plugin/list")
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))
}

func TestPluginAdd(t *testing.T) {
	postData := struct {
		Path   string            `json:"path"`
		Unique string            `json:"unique"`
		Params map[string]string `json:"params"`
	}{
		"show-time",
		"plugin-1",
		map[string]string{
			"fontfile": "/tmp/font.ttf",
		},
	}
	postDataBytes, _ := json.Marshal(postData)
	req, err := http.NewRequest("POST", Host+"plugin/add", bytes.NewBuffer(postDataBytes))
	assertError(t, err)

	resp, err := getClient().Do(req)
	assertError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assertError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))
}
