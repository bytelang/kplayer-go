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

// resource
func TestResourceAdd(t *testing.T) {
	postData := struct {
		Path   string `json:"path"`
		Unique string `json:"unique"`
	}{
		"/tmp/short.flv",
		"resource-1",
	}
	postDataBytes, _ := json.Marshal(postData)
	req, err := http.NewRequest("POST", Host+"resource/add", bytes.NewBuffer(postDataBytes))
	assertError(t, err)

	resp, err := getClient().Do(req)
	assertError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assertError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))

	// validate
	{
		req, err := http.NewRequest("GET", Host+"resource/list", nil)
		assertError(t, err)

		resp, err := getClient().Do(req)
		assertError(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		assertError(t, err)

		if resp.StatusCode != http.StatusOK {
			t.Fatal(string(body))
		}

		t.Log(string(body))

		resourcePath := gjson.Parse(string(body)).Get("resources.0.path").String()
		if resourcePath != "/tmp/short.flv" {
			t.Fatal("add resource failed")
		}
	}

	// remove
	{
		req, err := http.NewRequest("DELETE", Host+"resource/remove/resource-1", nil)
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

	// validate
	{
		req, err := http.NewRequest("GET", Host+"resource/list", nil)
		assertError(t, err)

		resp, err := getClient().Do(req)
		assertError(t, err)

		body, err := ioutil.ReadAll(resp.Body)
		assertError(t, err)

		if resp.StatusCode != http.StatusOK {
			t.Fatal(string(body))
		}

		t.Log(string(body))

		resourcePath := gjson.Parse(string(body)).Get("resources.0.path").String()
		if resourcePath == "/tmp/short.flv" {
			t.Fatal("add resource failed")
		}
	}
}

func TestResourceSeekStart(t *testing.T) {
	uniqueName, preSeek := getCurrentResourceSeek(t)
	if preSeek <= 5 {
		time.Sleep(time.Second * 5)
	}

	postData := struct {
		Seek   int64  `json:"seek"`
		Unique string `json:"unique"`
	}{
		0,
		uniqueName,
	}
	postDataBytes, _ := json.Marshal(postData)
	req, err := http.NewRequest("POST", Host+"resource/seek", bytes.NewBuffer(postDataBytes))
	assertError(t, err)

	resp, err := getClient().Do(req)
	assertError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assertError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))

	time.Sleep(time.Second * 3)

	// validate
	{
		queryUniqueName, seek := getCurrentResourceSeek(t)
		if queryUniqueName != uniqueName {
			t.Fatal("resource query not match")
		}

		if seek > 5 {
			t.Fatalf("seek failed. seek: %d", seek)
		}

		t.Logf("seek seconds: %d", seek)
	}
}

func TestResourceSeekDuration(t *testing.T) {
	uniqueName, preSeek := getCurrentResourceSeek(t)
	if preSeek <= 5 {
		time.Sleep(time.Second * 5)
	}

	postData := struct {
		Seek   int64  `json:"seek"`
		Unique string `json:"unique"`
	}{
		0,
		uniqueName,
	}
	postDataBytes, _ := json.Marshal(postData)
	req, err := http.NewRequest("POST", Host+"resource/seek", bytes.NewBuffer(postDataBytes))
	assertError(t, err)

	resp, err := getClient().Do(req)
	assertError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assertError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	t.Log(string(body))

	// validate
	{
		queryUniqueName, seek := getCurrentResourceSeek(t)
		if queryUniqueName != uniqueName {
			t.Fatal("resource query not match")
		}

		if seek > 5 {
			t.Fatal("seek failed")
		}
	}
}
