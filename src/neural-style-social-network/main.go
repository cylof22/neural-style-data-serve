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

	"neural-style-util"
)

var (
	serverPort   = flag.String("port", "8002", "neural style server port")
	consulAddr   = flag.String("consulAddr", "localhost", "consul service address")
	consulPort   = flag.String("consulPort", "8500", "consul service port")
	dbServerURL  = flag.String("dbserver", "apc-chain.documents.azure.com", "Mongodb server host")
	dbServerPort = flag.String("dbport", "10255", "Mongodb server port")
	dbUser       = flag.String("dbUser", "", "Mongodb user")
	dbKey        = flag.String("dbPassword", "", "Mongodb password")
	localDev     = flag.Bool("local", false, "Disable Cloud Storage and local Memcached")
)

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	reviews := session.DB("store").C("reviews")

	index := mgo.Index{
		Key:        []string{"productid"},
		Unique:     false,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := reviews.EnsureIndex(index)
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

	if *localDev {
		advertiseAddr = "localhost"
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

	// Register Social Service to Consul
	registar := NSUtil.Register(*consulAddr,
		*consulPort,
		advertiseAddr,
		*serverPort, "social", "v1", logger)

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
