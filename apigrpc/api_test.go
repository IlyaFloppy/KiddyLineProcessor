package apigrpc

import (
	"go.uber.org/zap"
	"kiddy-line-processor/db"
	"testing"
	"time"
)

func TestGRPCAPIStartStop(t *testing.T) {
	logger, _ := zap.NewProduction()

	repo := db.NewSportsPointsMockRepo()
	time.Sleep(time.Second * 4)

	Start(repo, "", "9090", logger)
	time.Sleep(time.Second)

	select {
	case err := <-Notify():
		t.Error(err)
	default:
	}

	Stop()
	time.Sleep(time.Second)

	select {
	case err := <-Notify():
		t.Error(err)
	default:
	}
}
