package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
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
		map[string]string{},
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

func TestPluginUpdate(t *testing.T) {
	time.Sleep(time.Second * 1)
	postData := struct {
		Unique string            `json:"unique"`
		Params map[string]string `json:"params"`
	}{
		"plugin-1",
		map[string]string{
			"fontcolor": "red",
		},
	}
	postDataBytes, _ := json.Marshal(postData)
	req, err := http.NewRequest("PATCH", Host+"plugin/update", bytes.NewBuffer(postDataBytes))
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

func TestPluginInvalidUpdate(t *testing.T) {
	postData := struct {
		Unique string            `json:"unique"`
		Params map[string]string `json:"params"`
	}{
		"invalid-1",
		map[string]string{
			"fontcolor": "red",
		},
	}
	postDataBytes, _ := json.Marshal(postData)
	req, err := http.NewRequest("PATCH", Host+"plugin/update", bytes.NewBuffer(postDataBytes))
	assertError(t, err)

	resp, err := getClient().Do(req)
	assertError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assertError(t, err)

	if resp.StatusCode == http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))
}

func TestPluginRemove(t *testing.T) {
	req, err := http.NewRequest("DELETE", Host+"plugin/remove/plugin-1", nil)
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

func TestPluginRemoveInvalid(t *testing.T) {
	req, err := http.NewRequest("DELETE", Host+"plugin/remove/plugin-1", nil)
	assertError(t, err)

	resp, err := getClient().Do(req)
	assertError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assertError(t, err)

	if resp.StatusCode == http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))
}
