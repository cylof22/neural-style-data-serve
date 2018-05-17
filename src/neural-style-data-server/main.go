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
	mgo "gopkg.in/mgo.v2"
)

var (
	serverURL               = flag.String("host", "0.0.0.0", "neural style server url")
	serverPort              = flag.String("port", "8000", "neural style server port")
	dbServerURL             = flag.String("dbserver", "0.0.0.0", "style products server url")
	dbServerPort            = flag.String("dbport", "9000", "style products port url")
	storageServerURL        = flag.String("storageURL", "0.0.0.0", "Storage Server URL")
	storageServerPort       = flag.String("storagePort", "5000", "Storage Server Port")
	storageServerSaveRouter = flag.String("saveRouter", "/api/v1/storage/save", "URL router for save")
	storageServerFindRouter = flag.String("findRouter", "/api/v1/storage/find", "URL router for find")
	cacheServer             = flag.String("cacheHost", "www.elforce.net", "memcached host")
	cacheGetRouter          = flag.String("cacheGetURL", "/api/v1/cache/get", "Cache Get Router")
	localDev                = flag.Bool("local", false, "Disable Cloud Storage and local Memcached")
	networkPath             = flag.String("network", "", "neural network model path")
	previewNetworkPath      = flag.String("previewNetwork", "", "neural network preview model path")
	outputPath              = flag.String("outputdir", "./", "neural style transfer output directory")
)

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	products := session.DB("store").C("products")

	index := mgo.Index{
		Key:        []string{"id"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := products.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	reviews := session.DB("store").C("reviews")

	index = mgo.Index{
		Key:        []string{"productId"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = reviews.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

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
	ensureIndex(session)

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stdout)
		logger = level.NewFilter(logger, level.AllowDebug())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	r := makeHTTPHandler(ctx, session, logger)

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
