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

	"neural-style-util"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/rs/cors"
	mgo "gopkg.in/mgo.v2"
)

var (
	serverPort              = flag.String("port", "8001", "neural style server port")
	consulAddr              = flag.String("consulAddr", "localhost", "consul service address")
	consulPort              = flag.String("consulPort", "8500", "consul service port")
	dbServerURL             = flag.String("dbserver", "0.0.0.0", "style products server url")
	dbServerPort            = flag.String("dbport", "9000", "style products port url")
	storageServerURL        = flag.String("storageURL", "0.0.0.0", "Storage Server URL")
	storageServerPort       = flag.String("storagePort", "5000", "Storage Server Port")
	storageServerSaveRouter = flag.String("saveRouter", "/api/v1/storage/save", "URL router for save")
	storageServerFindRouter = flag.String("findRouter", "/api/v1/storage/find", "URL router for find")
	cacheServer             = flag.String("cacheHost", "www.elforce.net", "memcached host")
	cacheGetRouter          = flag.String("cacheGetURL", "/api/v1/cache/get", "Cache Get Router")
	localDev                = flag.Bool("local", false, "Disable Cloud Storage and local Memcached")
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
}

func main() {
	flag.Parse()
	advertiseAddr, err := NSUtil.GetIPv4Host()
	if err != nil {
		fmt.Println(err.Error())
		return
	}

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
	r = cors.AllowAll().Handler(r)

	// Register Social Service to Consul
	registar := NSUtil.Register(*consulAddr,
		*consulPort,
		advertiseAddr,
		*serverPort, "products", logger)

	serverLoopBackURL := "0.0.0.0"
	// HTTP transport
	go func() {
		// How to show the debug info
		level.Debug(logger).Log("info", "Start server at port "+serverLoopBackURL+":"+*serverPort,
			"time", time.Now())
		// register service
		registar.Register()
		handler := r
		errChan <- http.ListenAndServe(serverLoopBackURL+":"+*serverPort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()

	errInfo := <-errChan
	registar.Deregister()
	defer func(end time.Time) {
		level.Debug(logger).Log("info", "End server: "+errInfo.Error(),
			"time", end)
	}(time.Now())
}
