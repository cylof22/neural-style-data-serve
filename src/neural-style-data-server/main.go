package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"neural-style-transfer"

	"github.com/go-kit/kit/log"
)

func main() {
	ctx := context.Background()
	errChan := make(chan error)

	var svc StyleService.Service
	svc = StyleService.NeuralTransferService{}

	endpoint := StyleService.Endpoints{
		NeuralStyleEndpoint: StyleService.MakeNeuralStyleEndpoint(svc),
	}

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	r := StyleService.MakeHTTPHandler(ctx, endpoint, logger)

	// HTTP transport
	go func() {
		fmt.Println("Starting server at port 9090")
		handler := r
		errChan <- http.ListenAndServe(":9090", handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	fmt.Println(<-errChan)
}
