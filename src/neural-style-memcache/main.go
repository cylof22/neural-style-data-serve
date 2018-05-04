package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-kit/kit/log"
)

var (
	cacheHost = flag.String("host", "localhost", "Memecached service host")
	cachePort = flag.String("port", "9999", "Memecached service port")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	r := makeHTTPHandler(ctx, logger)

	// HTTP transport
	go func() {
		fmt.Println("Starting server at port " + *cachePort)
		handler := r
		errChan <- http.ListenAndServe(*cacheHost+":"+*cachePort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	fmt.Println(<-errChan)

}
