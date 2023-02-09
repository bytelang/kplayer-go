package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"
)

// output
func TestOutputList(t *testing.T) {
	resp, err := getClient().Get(Host + "output/list")
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

func TestOutputAddInvalid(t *testing.T) {
	postData := struct {
		Path   string `json:"path"`
		Unique string `json:"unique"`
	}{
		"/invalid",
		"invalid-1",
	}
	postDataBytes, _ := json.Marshal(postData)

	req, err := http.NewRequest("POST", Host+"output/add", bytes.NewBuffer(postDataBytes))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := getClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// invalid output should not 200
	if resp.StatusCode == http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))
}

func TestOutputAddValid(t *testing.T) {
	postData := struct {
		Path   string `json:"path"`
		Unique string `json:"unique"`
	}{
		"output-1",
		"valid-1",
	}
	postDataBytes, _ := json.Marshal(postData)

	req, err := http.NewRequest("POST", Host+"output/add", bytes.NewBuffer(postDataBytes))
	if err != nil {
		t.Fatal(err)
	}

	resp, err := getClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// invalid output should not 200
	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))
}

func TestOutputRemove(t *testing.T) {
	req, err := http.NewRequest("DELETE", Host+"output/remove/valid-1", nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := getClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// invalid output should not 200
	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))
}
