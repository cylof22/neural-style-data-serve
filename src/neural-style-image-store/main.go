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
	mgo "gopkg.in/mgo.v2"
)

var (
	serverURL    = flag.String("host", "0.0.0.0", "neural style server url")
	serverPort   = flag.String("port", "5000", "neural style server port")
	dbServerURL  = flag.String("dbserver", "apc-chain.documents.azure.com", "Mongodb server host")
	dbServerPort = flag.String("dbport", "10255", "Mongodb server port")
	dbUser       = flag.String("dbUser", "", "Mongodb user")
	dbKey        = flag.String("dbPassword", "", "Mongodb password")
)

func main() {
	flag.Parse()
	errChan := make(chan error)

	ctx := context.Background()
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

	// Logging domain.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		level.NewFilter(logger, level.AllowDebug())
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	r := makeHTTPHandler(ctx, session, logger)

	// HTTP transport
	go func() {
		level.Debug(logger).Log("info", "Starting server at "+*serverURL+":"+*serverPort)
		handler := r
		errChan <- http.ListenAndServe(*serverURL+":"+*serverPort, handler)
	}()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
	errinfo := <-errChan
	level.Debug(logger).Log("info", "crash info: "+errinfo.Error())
}
