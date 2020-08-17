package apigrpc

import (
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"kiddy-line-processor/apigrpc/gen"
	"kiddy-line-processor/db"
	"net"
)

var (
	errorsChan chan error
	grpcServer *grpc.Server
	slogger    *zap.SugaredLogger
)

// Start start gRPC API server
// Does not block
func Start(repo db.SportsPointsRepo, grpcAddress, grpcPort string, logger *zap.Logger) {
	errorsChan = make(chan error, 1)

	go func() {
		slogger = logger.Sugar()
		slogger.Debug("Starting grpc api server")
		lis, err := net.Listen("tcp", net.JoinHostPort(grpcAddress, grpcPort))
		if err != nil {
			slogger.Errorw("Failed to listen (gRPC API)", "err", err)
			errorsChan <- err
		}
		grpcServer = grpc.NewServer()
		gen.RegisterSportsLinesServer(grpcServer, &sportsLinesGrpcImpl{
			repo:    repo,
			slogger: slogger,
		})

		err = grpcServer.Serve(lis)
		if err != nil {
			slogger.Errorw("Error accepting connection (gRPC API)", "err", err)
			errorsChan <- err
		}
	}()
}

// Notify returns a channel to notify caller about errors
func Notify() <-chan error {
	return errorsChan
}

// Stop stops grpc api server
func Stop() {
	slogger.Debug("Stopping grpc api server")
	grpcServer.GracefulStop()
}
