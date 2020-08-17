package apihttp

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

type status struct {
	Status string `json:"status"`
}

func TestHTTPAPIStartStop(t *testing.T) {
	logger, _ := zap.NewProduction()

	Start("", "8080", logger)

	var stat status
	response, _ := http.Get("http://localhost:8080/ready")
	data, _ := ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(data, &stat)
	if stat.Status != "not ok" {
		t.Fail()
	}

	SetReady(true)

	response, _ = http.Get("http://localhost:8080/ready")
	data, _ = ioutil.ReadAll(response.Body)
	_ = json.Unmarshal(data, &stat)
	if stat.Status != "ok" {
		t.Fail()
	}

	select {
	case err := <-Notify():
		t.Error(err)
	default:
	}

	err := Stop(time.Second)
	if err != nil {
		t.Error(err)
	}

	select {
	case err := <-Notify():
		t.Error(err)
	default:
	}
}
