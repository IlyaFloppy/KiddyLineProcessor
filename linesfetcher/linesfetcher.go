package linesfetcher

import (
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"kiddy-line-processor/db"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	errorsChan                 chan error
	sportsNotSyncedAtLeastOnce sync.Map
	client                     *http.Client
	addr                       string
	prt                        string
	stopped                    bool
	slogger                    *zap.SugaredLogger
)

type linesProviderResponse struct {
	Lines map[string]string `json:"lines"`
}

func updateSport(repo db.SportsPointsRepo, sport string) error {
	url := fmt.Sprintf("http://%s:%s/api/v1/lines/%s", addr, prt, sport)
	response, err := client.Get(url)
	if err != nil {
		return err
	}

	buf, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	var lines linesProviderResponse
	err = json.Unmarshal(buf, &lines)
	if err != nil {
		return err
	}

	if valueString, ok := lines.Lines[strings.ToUpper(sport)]; ok {
		value, err := strconv.ParseFloat(valueString, 32)
		if err != nil {
			return err
		}
		_ = repo.PutPoint(sport, db.Point{
			Time:  time.Now(),
			Value: float32(value),
		})
	} else {
		return errors.New("failed to parse response")
	}

	return nil
}

// Start starts lines fetcher
// Blocks until
func Start(
	repo db.SportsPointsRepo,
	address, port string,
	sportUpdateIntervals map[string]time.Duration,
	logger *zap.Logger) {

	addr = address
	prt = port

	slogger = logger.Sugar()
	slogger.Debug("Starting fetching lines from lines provider")

	stopped = false

	errorsChan = make(chan error)

	client = &http.Client{
		Timeout: time.Second * 3,
	}

	for s, i := range sportUpdateIntervals {
		sport := s
		interval := i
		sportsNotSyncedAtLeastOnce.Store(sport, true)
		go func() {
			for {
				if stopped {
					return
				}
				err := updateSport(repo, sport)
				if err != nil {
					slogger.Warnw("Failed to update sport from lines provider", "err", err, "sport", sport)
				} else {
					sportsNotSyncedAtLeastOnce.Delete(sport)
				}
				time.Sleep(interval)
			}
		}()
	}
}

// Notify returns a channel to notify caller about errors
func Notify() <-chan error {
	return errorsChan
}

// Ready checks if each sport was synchronized at least once
func Ready() bool {
	notSyncedSportsCount := 0
	sportsNotSyncedAtLeastOnce.Range(func(key, value interface{}) bool {
		notSyncedSportsCount += 1
		return false
	})
	return notSyncedSportsCount == 0
}

// Stop stops lines fetcher
func Stop() {
	stopped = true
}
