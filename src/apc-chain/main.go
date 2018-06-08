package main

import (
	"flag"
	"fmt"
	"net/http"

	mgo "gopkg.in/mgo.v2"
)

var (
	serverURL    = flag.String("host", "0.0.0.0", "neural style server url")
	serverPort   = flag.String("port", "8000", "neural style server port")
	dbServerURL  = flag.String("dbserver", "0.0.0.0", "style products server url")
	dbServerPort = flag.String("dbport", "9000", "style products port url")
)

func main() {
	session, err := mgo.Dial(*dbServerURL + ":" + *dbServerPort)

	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	homeServer := http.FileServer(http.Dir("dist"))
	http.Handle("/", homeServer)

	docServer := http.FileServer(http.Dir("documents"))
	docHandler := http.StripPrefix("/documents/", docServer)
	http.Handle("/documents/", docHandler)

	// add tencent yun authorization ssl file
	authFiles := http.FileServer(http.Dir("data/auth/"))
	http.StripPrefix("/.well-known/pki-validation/", authFiles)

	tokensvc := NewTokenPreSale(session)
	http.Handle("/token", tokensvc)

	err = http.ListenAndServe(*serverURL+":"+*serverPort, nil)
	if err != nil {
		fmt.Println(err)
	}

}
