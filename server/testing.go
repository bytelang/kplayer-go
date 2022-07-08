package server

import (
	"github.com/tidwall/gjson"
	"io/ioutil"
	"net"
	"net/http"
	"testing"
	"time"
)

const Host = "http://127.0.0.1:4156/"

func assertError(t *testing.T, err error, msg ...string) {
	if err != nil {
		if len(msg) == 0 {
			t.Fatal(err)
		} else {
			t.Fatal(msg[0])
		}
	}
}

func getClient() *http.Client {
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client
}

func getCurrentResourceSeek(t *testing.T) (string, int64) {
	// query current duration
	req, err := http.NewRequest("GET", Host+"resource/current", nil)
	if err != nil {
		assertError(t, err)
	}

	resp, err := getClient().Do(req)
	assertError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	assertError(t, err)

	if resp.StatusCode != http.StatusOK {
		t.Fatal(string(body))
	}

	unique := gjson.Parse(string(body)).Get("resource.unique").String()
	seek := gjson.Parse(string(body)).Get("seek").Int()

	return unique, seek
}
