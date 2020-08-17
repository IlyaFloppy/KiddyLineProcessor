package apihttp

import (
	"context"
	"github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"kiddy-line-processor/apihttp/handler"
	"net"
	"net/http"
	"time"
)

var (
	errorsChan chan error
	server     *http.Server
	slogger    *zap.SugaredLogger
	ready      bool
)

// Start starts http api server on given address
// Does not block
func Start(httpAddress, httpPort string, logger *zap.Logger) {
	slogger = logger.Sugar()
	slogger.Info("Http api server is starting")
	errorsChan = make(chan error, 1)

	router := gin.Default()

	router.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(logger, true))

	router.GET("/ready", handler.ReadyGet(&ready))

	server = &http.Server{
		Addr:    net.JoinHostPort(httpAddress, httpPort),
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			slogger.Errorw("Http api error occurred", "err", err)
			errorsChan <- err
		}
	}()
	slogger.Info("Http api server is started")
}

// Notify returns a channel to notify caller about errors
func Notify() <-chan error {
	return errorsChan
}

// SetReady set ready status
func SetReady(r bool) {
	ready = r
}

// Stop stops http api server
func Stop(timeout time.Duration) error {
	slogger.Info("Http api server is shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		slogger.Errorw("Failed to shutdown http api server", "err", err)
		return err
	}
	slogger.Info("Http api server is shut down")
	return nil
}
