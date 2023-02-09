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
		"short.flv",
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
		if resourcePath != "short.flv" {
			t.Fatal("add resource failed")
		}
	}

	// remove
	defer removeResource("resource-1", t)

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
		if resourcePath == "short.flv" {
			t.Fatal("add resource failed", string(body))
		}
	}
}

// resource
func TestMixResourceAdd(t *testing.T) {
	postData := struct {
		Unique          string                   `json:"unique"`
		MixResourceType bool                     `json:"mix_resource_type"`
		Groups          []map[string]interface{} `json:"groups"`
	}{
		"resource-1",
		true,
		[]map[string]interface{}{
			{
				"path":       "short.flv",
				"media_type": "video",
			},
			{
				"path":       "short.flv",
				"media_type": "audio",
			},
		},
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
		if resourcePath != "short.flv" {
			t.Fatal("add resource failed")
		}
	}

	// remove
	removeResource("resource-1", t)

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
		if resourcePath == "short.flv" {
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

func TestMixResourceSeekUnique(t *testing.T) {
	defer removeResource("resource-1", t)
	{
		postData := struct {
			Unique          string                   `json:"unique"`
			MixResourceType bool                     `json:"mix_resource_type"`
			Groups          []map[string]interface{} `json:"groups"`
		}{
			"resource-1",
			true,
			[]map[string]interface{}{
				{
					"path":       "short.flv",
					"media_type": "video",
				},
				{
					"path":       "short.flv",
					"media_type": "audio",
				},
			},
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
	}

	postData := struct {
		Seek   int64  `json:"seek"`
		Unique string `json:"unique"`
	}{
		0,
		"resource-1",
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
		if queryUniqueName != "resource-1" {
			t.Fatal("resource query not match")
		}

		t.Logf("seek seconds: %d", seek)
	}
}

func TestMixResourceSeekStart(t *testing.T) {
	TestMixResourceSeekUnique(t)
	defer func() {
		removeResource("resource-1", t)
	}()

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

func TestMixResourceSeekDuration(t *testing.T) {
	TestMixResourceSeekUnique(t)

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
	time.Sleep(time.Second * 3)
	{
		queryUniqueName, seek := getCurrentResourceSeek(t)
		if queryUniqueName != uniqueName {
			t.Fatal("resource query not match")
		}

		if seek > 5 {
			t.Fatalf("seek failed. seek: %d", seek)
		}
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
	time.Sleep(time.Second * 3)
	{
		queryUniqueName, seek := getCurrentResourceSeek(t)
		if queryUniqueName != uniqueName {
			t.Fatal("resource query not match")
		}

		if seek > 5 {
			t.Fatalf("seek failed. seek: %d", seek)
		}
	}
}

func removeResource(unique string, t *testing.T) {
	// remove
	{
		req, err := http.NewRequest("DELETE", Host+"resource/remove/"+unique, nil)
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
}
