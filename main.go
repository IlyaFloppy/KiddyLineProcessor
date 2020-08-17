package main

import (
	"encoding/json"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"kiddy-line-processor/apigrpc"
	"kiddy-line-processor/apihttp"
	"kiddy-line-processor/db"
	"kiddy-line-processor/linesfetcher"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	postgresHost     string
	postgresPort     string
	postgresUser     string
	postgresPassword string
	postgresName     string

	apiHTTPAddress string
	apiHTTPPort    string
	apiGrpcAddress string
	apiGrpcPort    string

	fetchAddress string
	fetchPort    string
	fetchSports  map[string]uint

	logLevel zapcore.Level
)

func init() {
	postgresHost = os.Getenv("POSTGRES_HOST")
	postgresPort = os.Getenv("POSTGRES_PORT")
	postgresUser = os.Getenv("POSTGRES_USER")
	postgresPassword = os.Getenv("POSTGRES_PASSWORD")
	postgresName = os.Getenv("POSTGRES_NAME")

	apiHTTPAddress = os.Getenv("KLP_HTTP_ADDRESS")
	apiHTTPPort = os.Getenv("KLP_HTTP_PORT")
	apiGrpcAddress = os.Getenv("KLP_GRPC_ADDRESS")
	apiGrpcPort = os.Getenv("KLP_GRPC_PORT")

	switch level := os.Getenv("KLP_LOG_LEVEL"); level {
	case "debug":
		logLevel = zapcore.DebugLevel
	case "info":
		logLevel = zapcore.InfoLevel
	case "warn":
		logLevel = zapcore.WarnLevel
	case "error":
		logLevel = zapcore.ErrorLevel
	case "panic":
		logLevel = zapcore.PanicLevel
	case "fatal":
		logLevel = zapcore.FatalLevel
	default:
		logLevel = zapcore.DebugLevel
	}

	fetchAddress = os.Getenv("FETCH_ADDRESS")
	fetchPort = os.Getenv("FETCH_PORT")
	err := json.Unmarshal([]byte(os.Getenv("FETCH_SPORTS")), &fetchSports)
	if err != nil {
		log.Fatal(os.Getenv("FETCH_SPORTS"), err)
	}
}

func main() {
	// create logger
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(logLevel)
	logger, err := config.Build()
	if err != nil {
		panic(err) // failed to initialize logger
	}
	defer func() { _ = logger.Sync() }()
	slogger := logger.Sugar()
	defer func() { _ = slogger.Sync() }()

	slogger.Info("Starting the application...")

	// create db interface
	repo := db.NewPGSportsPointsRepo(slogger)

	// start http api
	apihttp.Start(apiHTTPAddress, apiHTTPPort, logger)

	// connect db and create required tables if they don't exist
	err = repo.Connect(postgresHost, postgresPort, postgresUser, postgresPassword, postgresName)
	if err != nil {
		slogger.Panic("Failed to connect to database. Panic.")
	}
	sports := make([]string, 0)
	for sport := range fetchSports {
		sports = append(sports, sport)
	}
	err = repo.CreateTablesIfDontExist(sports...)
	if err != nil {
		slogger.Panic("Failed to create tables. Panic.")
	}

	// start fetching lines from lines provider and wait for first synchronization
	sportUpdatesIntervals := make(map[string]time.Duration)
	for sport, interval := range fetchSports {
		sportUpdatesIntervals[sport] = time.Second * time.Duration(interval)
	}
	linesfetcher.Start(repo, fetchAddress, fetchPort, sportUpdatesIntervals, logger)
	for !linesfetcher.Ready() {
		time.Sleep(time.Millisecond * 500)
	}

	// start grpc api
	apigrpc.Start(repo, apiGrpcAddress, apiGrpcPort, logger)

	apihttp.SetReady(true)

	// catch signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// exit on errors/signals
errorsWaitLoop:
	for {
		select {
		case sig := <-signals:
			slogger.Infow("Application stopped (system signal)", "signal", sig.String())
			break errorsWaitLoop
		case err = <-repo.Notify():
			slogger.Errorw("Error in database", "err", err)
			break errorsWaitLoop
		case err = <-apihttp.Notify():
			slogger.Errorw("Error in http api", "err", err)
		case err = <-apigrpc.Notify():
			slogger.Errorw("Error in grpc api", "err", err)
		case err = <-linesfetcher.Notify():
			slogger.Errorw("Error fetching lines", "err", err)
		}
	}

	_ = apihttp.Stop(time.Second * 3)
	linesfetcher.Stop()
	apigrpc.Stop()
	_ = repo.Close()
}
