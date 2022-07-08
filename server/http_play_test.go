package server

import (
	"bytes"
	"encoding/json"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

// play
func TestPlayDuration(t *testing.T) {
	resp, err := getClient().Get(Host + "play/duration")
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

func TestPlayPause(t *testing.T) {
	// get current duration
	_, preSeek := getCurrentResourceSeek(t)

	req, err := http.NewRequest("POST", Host+"play/pause", nil)
	assertError(t, err)

	resp, err := getClient().Do(req)
	assertError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assertError(t, err)

	// invalid output should not 200
	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	// validate seek
	time.Sleep(time.Second * 5)
	_, pauseSeek := getCurrentResourceSeek(t)

	if pauseSeek-preSeek > 3 {
		t.Fatal("pause failed")
	}

	t.Log(string(body))

	// continue
	{
		req, err := http.NewRequest("POST", Host+"play/continue", nil)
		assertError(t, err)

		resp, err := getClient().Do(req)
		assertError(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		assertError(t, err)

		// invalid output should not 200
		if resp.StatusCode != http.StatusOK {
			t.Fatal(string(body))
		}

		t.Log(string(body))
	}
}

func TestPlayInformation(t *testing.T) {
	req, err := http.NewRequest("GET", Host+"play/information", nil)
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

func TestPlaySkip(t *testing.T) {
	// add resource
	{
		postData := struct {
			Path   string `json:"path"`
			Unique string `json:"unique"`
		}{
			"short.flv",
			"resource-1",
		}
		postDataBytes, _ := json.Marshal(postData)
		req, err := http.NewRequest("POST", Host+"resource/add", bytes.NewBuffer(postDataBytes))
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

		if resp.StatusCode != http.StatusOK {
			t.Fatal(string(body))
		}

		t.Log(string(body))
	}

	// skip
	req, err := http.NewRequest("POST", Host+"play/skip", nil)
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

	// wait a moment
	time.Sleep(time.Second * 5)
	// validate resource
	{
		req, err := http.NewRequest("GET", Host+"resource/current", nil)
		assertError(t, err)

		resp, err := getClient().Do(req)
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

		unique := gjson.Parse(string(body)).Get("resource.unique").String()
		if unique != "resource-1" {
			t.Fatal(string(body))
		}

		seek := gjson.Parse(string(body)).Get("seek").String()
		if seek == "0" {
			t.Fatalf("resource seek state invalid. seek: %s", seek)
		}

		t.Log(string(body))
	}
}

func TestPlayStop(t *testing.T) {
	req, err := http.NewRequest("POST", Host+"play/stop", nil)
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
