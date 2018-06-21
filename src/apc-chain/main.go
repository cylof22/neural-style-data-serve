package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/http"
	"time"

	mgo "gopkg.in/mgo.v2"
)

var (
	serverURL    = flag.String("host", "0.0.0.0", "neural style server url")
	serverPort   = flag.String("port", "80", "neural style server port")
	dbServerURL  = flag.String("dbserver", "apc-chain.documents.azure.com", "Mongodb server host")
	dbServerPort = flag.String("dbport", "10255", "Mongodb server port")
	dbUser       = flag.String("dbUser", "", "Mongodb user")
	dbKey        = flag.String("dbPassword", "", "Mongodb password")
	local        = flag.Bool("local", false, "flag for store local file")
)

func ensureIndex(s *mgo.Session) {
	session := s.Copy()
	defer session.Close()

	tokens := session.DB("store").C("tokens")

	index := mgo.Index{
		Key:        []string{"address"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}
	err := tokens.EnsureIndex(index)
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.Parse()

	dbAddr := *dbServerURL + ":" + *dbServerPort
	if *local {
		dbAddr = "0.0.0.0:9000"
	}

	dialInfo := &mgo.DialInfo{
		Addrs:    []string{dbAddr},
		Timeout:  10 * time.Second,
		Database: "store",
	}

	if !(*local) {
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

	homeServer := http.FileServer(http.Dir("dist"))
	http.Handle("/", homeServer)
	enServer := http.StripPrefix("/en/", homeServer)
	http.Handle("/en/", enServer)

	airdropServer := http.StripPrefix("/ch/airdrop/", homeServer)
	http.Handle("/ch/airdrop/", airdropServer)

	docServer := http.FileServer(http.Dir("documents"))
	docHandler := http.StripPrefix("/documents/", docServer)
	http.Handle("/documents/", docHandler)

	tokensvc := NewTokenPreSale(session)
	http.Handle("/token", tokensvc)

	picturesvc := NewPictureService(session)
	http.Handle("/picture", picturesvc)

	err = http.ListenAndServe(*serverURL+":"+*serverPort, nil)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("APC-chain Server start: ", *serverURL+":"+*serverPort)
}
