package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/rs/cors"
	mgo "gopkg.in/mgo.v2"
)

var (
	serverURL               = flag.String("host", "0.0.0.0", "neural style server url")
	serverPort              = flag.String("port", "8000", "neural style server port")
	dbServerURL             = flag.String("dbserver", "apc-chain.documents.azure.com", "Mongodb server host")
	dbServerPort            = flag.String("dbport", "10255", "Mongodb server port")
	dbUser                  = flag.String("dbUser", "", "Mongodb user")
	dbKey                   = flag.String("dbPassword", "", "Mongodb password")
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
	productsRouter          = flag.String("productsRouter", "/api/products", "URL router for products")
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

	orders := session.DB("store").C("orders")
	index = mgo.Index{
		Key:        []string{"productId"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err = orders.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	ctx := context.Background()
	errChan := make(chan error)

	dbAddr := *dbServerURL + ":" + *dbServerPort
	if *localDev {
		dbAddr = "0.0.0.0:9000"
	}

	dialInfo := &mgo.DialInfo{
		Addrs:    []string{dbAddr},
		Timeout:  10 * time.Second,
		Database: "store",
	}

	if !(*localDev) {
		dialInfo.Username = *dbUser
		dialInfo.Password = *dbKey
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			return tls.Dial("tcp", addr.String(), &tls.Config{})
		}
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		fmt.Println("Db connection fails: " + err.Error())
		return
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
