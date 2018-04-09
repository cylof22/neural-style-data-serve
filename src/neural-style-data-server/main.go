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
	mgo "gopkg.in/mgo.v2"
)

var serverURL = flag.String("host", "localhost", "neural style server url")
var serverPort = flag.String("port", "9090", "neural style server port")
var dbServerURL = flag.String("dbserver", "localhost", "style products server url")
var dbServerPort = flag.String("dbport", "9000", "style products port url")
var networkPath = flag.String("network", "", "neural network model path")
var previewNetworkPath = flag.String("previewNetwork", "", "neural network preview model path")
var outputPath = flag.String("outputdir", "./", "neural style transfer output directory")

func main() {
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	session, err := mgo.Dial(*dbServerURL + ":" + *dbServerPort)
	if err != nil {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	var svc StyleService.Service
	svc = StyleService.NeuralTransferService{
		NetworkPath:        *networkPath,
		PreviewNetworkPath: *previewNetworkPath,
		OutputPath:         *outputPath,
		Host:               *serverURL,
		Port:               *serverPort,
		Session:            session,
	}

	endpoint := StyleService.Endpoints{
		NSEndpoint:                StyleService.MakeNSEndpoint(svc),
		NSPreviewEndpoint:         StyleService.MakeNSPreviewEndpoint(svc),
		NSContentUploadEndpoint:   StyleService.MakeNSContentUploadEndpoint(svc),
		NSStyleUploadEndpoint:     StyleService.MakeNSStyleUploadEndpoint(svc),
		NSGetProductsEndpoint:     StyleService.MakeNSGetProductsEndpoint(svc),
		NSGetProductsByIDEndpoint: StyleService.MakeNSGetProductByIDEndpoint(svc),
		NSGetReviewsByIDEndpoint:  StyleService.MakeNSGetReviewsByIDEndpoint(svc),
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
