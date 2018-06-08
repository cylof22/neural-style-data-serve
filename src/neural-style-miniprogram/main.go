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
	serverURL    = flag.String("host", "0.0.0.0", "neural style server url")
	serverPort   = flag.String("port", "9090", "neural style server port")
	domainURL    = flag.String("domain", "tulian.17dodo.com", "external domain")
	domainPort   = flag.String("dPort", "9090", "external domain port")
	certFile     = flag.String("cert", "./data/tls/214699506910084.pem", "TLS cert file path")
	keyFile      = flag.String("key", "./data/tls/214699506910084.key", "TLS key file path")
	transferURL  = flag.String("aihost", "39.106.123.1", "AI Server")
	transferPort = flag.String("aiport", "9091", "AI Port")
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
		errChan <- http.ListenAndServeTLS(*serverURL+":"+*serverPort, *certFile, *keyFile, handler)
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
