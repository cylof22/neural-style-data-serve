package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"neural-style-transfer"

	"github.com/go-kit/kit/log"
)

var serverURL = flag.String("host", "localhost", "neural style server url")
var serverPort = flag.String("port", "9090", "neural style server port")
var networkPath = flag.String("network", "", "neural network model path")

func main() {
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	var svc StyleService.Service
	svc = StyleService.NeuralTransferService{
		NetworkPath: *networkPath,
	}

	endpoint := StyleService.Endpoints{
		NeuralStyleEndpoint:        StyleService.MakeNeuralStyleEndpoint(svc),
		NeuralStylePreviewEndpoint: StyleService.MakeNeuralStylePreviewEndpoint(svc),
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
		errChan <- http.ListenAndServe(*serverURL+":"+*serverPort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	fmt.Println(<-errChan)
}
