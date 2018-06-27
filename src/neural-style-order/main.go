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

	"neural-style-util"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/rs/cors"

	mgo "gopkg.in/mgo.v2"
)

var (
	serverPort   = flag.String("port", "8004", "neural style server port")
	consulAddr   = flag.String("consulAddr", "localhost", "consul service address")
	consulPort   = flag.String("consulPort", "8500", "consul service port")
	dbServerURL  = flag.String("dbserver", "apc-chain.documents.azure.com", "Mongodb server host")
	dbServerPort = flag.String("dbport", "10255", "Mongodb server port")
	dbUser       = flag.String("dbUser", "", "Mongodb user")
	dbKey        = flag.String("dbPassword", "", "Mongodb password")
	productsURL  = flag.String("productsURL", "/api/products", "URL router for products")
	localDev     = flag.Bool("local", false, "local host debug environment")
	apiSite      = flag.String("apiSite", "localhost:8000", "api service site")
)

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	products := session.DB("store").C("order")

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

	dbAddr := *dbServerURL + ":" + *dbServerPort

	dialInfo := &mgo.DialInfo{
		Addrs:    []string{dbAddr},
		Timeout:  10 * time.Second,
		Database: "store",
	}

	dialInfo.Username = *dbUser
	dialInfo.Password = *dbKey
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), &tls.Config{})
	}

	session, err := mgo.DialWithInfo(dialInfo)
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

	// Register User Service to Consul
	registar := NSUtil.Register(*consulAddr,
		*consulPort,
		advertiseAddr,
		*serverPort, "orders", "v1", logger)

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
