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
	local        = flag.Bool("debug", false, "flag for store local file")
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
	session, err := mgo.Dial(*dbServerURL + ":" + *dbServerPort)

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

}
