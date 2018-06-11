package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/rs/cors"
)

var (
	serverURL      = flag.String("host", "0.0.0.0", "neural style server url")
	serverPort     = flag.String("port", "8000", "neural style server port")
	prodURL        = flag.String("prodURL", "0.0.0.0", "neural style product server url")
	prodPort       = flag.String("prodPort", "8000", "neural style product port")
	productsRouter = flag.String("productsRouter", "/api/products", "URL router for products")
)

func main() {
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = level.NewFilter(logger, level.AllowDebug())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	db, err := sql.Open("postgres", "user=Arnold dbname=TotalRecall sslmode=disable")
	if err != nil {
		level.Debug(logger).Log("Info", "Unable to Open the database", "Error", err.Error())
		return
	}

	r := makeHTTPHandler(ctx, db, logger)

	r = cors.Default().Handler(r)

	// HTTP transport
	go func() {
		// How to show the debug info
		level.Debug(logger).Log("info", "Start server at port "+*serverURL+":"+*serverPort,
			"time", time.Now())
		handler := r
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
