package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

var (
	serverURL    = flag.String("host", "localhost", "neural style server url")
	serverPort   = flag.String("port", "9090", "neural style server port")
	certFile     = flag.String("cert", "", "TLS cert file path")
	keyFile      = flag.String("key", "", "TLS key file path")
	transferURL  = flag.String("aihost", "localhost", "AI Server")
	transferPort = flag.String("aiport", "5000", "AI Port")
)

func main() {
	ctx := context.Background()
	errChan := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = level.NewFilter(logger, level.AllowDebug())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	r := MakeHTTPHandler(ctx)

	// HTTP transport
	go func() {
		// How to show the debug info
		level.Debug(logger).Log("info", "Start server at port "+*serverURL+":"+*serverPort,
			"time", time.Now())
		handler := r
		//errChan <- http.ListenAndServeTLS(*serverURL+":"+*serverPort, *certFile, *keyFile, handler)
		errChan <- http.ListenAndServe(*serverURL+":"+*serverPort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	errInfo := <-errChan
	defer func(end time.Time) {
		level.Debug(logger).Log("info", "End server: "+errInfo.Error(),
			"time", end)
	}(time.Now())
}
