package apigrpc

import (
	"errors"
	"go.uber.org/zap"
	"io"
	"kiddy-line-processor/apigrpc/gen"
	"kiddy-line-processor/db"
	"time"
)

type sportsLinesGrpcImpl struct {
	repo    db.SportsPointsRepo
	slogger *zap.SugaredLogger
}

type requirementsChange struct {
	interval time.Duration
	sports   []string
}

func (g *sportsLinesGrpcImpl) sendDeltasAndGetNewValues(
	stream gen.SportsLines_SubscribeOnSportsLinesServer,
	previousValues map[string]float32) (map[string]float32, error) {

	response := gen.SubscriptionResponse{Deltas: make(map[string]float32)}
	newValues := make(map[string]float32)
	for sport := range previousValues {
		point, err := g.repo.GetCurrent(sport)
		if err != nil {
			return previousValues, err
		}
		newValues[sport] = point.Value
		response.Deltas[sport] = newValues[sport] - previousValues[sport]
	}

	err := stream.Send(&response)
	if err != nil {
		return previousValues, err
	}

	return newValues, nil
}

func mapKeysEqual(m map[string]float32, keys []string) bool {
	if len(m) == len(keys) {
		for _, key := range keys {
			if _, ok := m[key]; !ok {
				return false
			}
		}
		return true
	}
	return false
}

func (g *sportsLinesGrpcImpl) sendDeltasWorker(
	stream gen.SportsLines_SubscribeOnSportsLinesServer,
	requirementsChan <-chan requirementsChange,
	finishChan <-chan bool) {

	ticker := time.NewTicker(time.Second)

	var values map[string]float32
	var err error
	for {
		select {
		case finish := <-finishChan:
			if finish {
				return
			}
		case requirements := <-requirementsChan:
			ticker = time.NewTicker(requirements.interval)
			// reset values and send full lines if sports changed
			if !mapKeysEqual(values, requirements.sports) {
				values = make(map[string]float32)
				for _, sport := range requirements.sports {
					values[sport] = 0
				}
			}
		case <-ticker.C:
			if len(values) == 0 {
				continue
			}
		}
		// either received new requirements or the ticker ticked
		values, err = g.sendDeltasAndGetNewValues(stream, values)
		if err != nil {
			g.slogger.Warnw("Error occurred while sending data to grpc client", "err", err)
		}
	}
}

func (g *sportsLinesGrpcImpl) SubscribeOnSportsLines(stream gen.SportsLines_SubscribeOnSportsLinesServer) error {
	ctx := stream.Context()

	requirementsChan := make(chan requirementsChange)
	finishChan := make(chan bool)

	go g.sendDeltasWorker(stream, requirementsChan, finishChan)

	for {
		select {
		case <-ctx.Done():
			finishChan <- true
			return ctx.Err()
		default:
		}

		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			g.slogger.Warnw("Error occurred while receiving grpc request", "err", err)
			finishChan <- true
			return err
		}
		if len(req.GetSports()) == 0 {
			g.slogger.Warn("Grpc client requested zero length list of sports")
		}
		if req.Interval == 0 {
			err := errors.New("grpc client tried to subscribe on sport lines with zero interval")
			g.slogger.Errorw("SubscribeOnSportsLines error", "err", err)
			return err
		}

		requirementsChan <- requirementsChange{
			interval: time.Second * time.Duration(req.GetInterval()),
			sports:   req.GetSports(),
		}
	}
}
