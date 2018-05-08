package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
)

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("context-type", "application/json,charset=utf8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

// makeHTTPHandler generate the http handler for storage service
func makeHTTPHandler(ctx context.Context, dbSession *mgo.Session, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	var svc Service
	svc = NewStorageService(dbSession)
	svc = NewLoggingService(log.With(logger, "component", "storage"), svc)

	//POST /api/v1/storage/save/{userid}/{imageid}
	r.Methods("POST").Path("/api/v1/storage/save").Queries("userid", "{userid}", "imageid", "{imageid}").Handler(
		httptransport.NewServer(
			MakeNSStoreEndpoint(svc),
			decodeNSSaveRequest,
			encodeNSSaveResponse,
			options...,
		))

	// GET /api/v1/storage/find/{userid}/{imageid}
	r.Methods("GET").Path("/api/v1/storage/find").Queries("userid", "{userid}", "imageid", "{imageid}").Handler(
		httptransport.NewServer(
			MakeNSFindEndpoint(svc),
			decodeNSFindRequest,
			encodeNSFindResponse,
			options...,
		))

	return r
}

// NSSaveRequest parse the basic information for uploading
type NSSaveRequest struct {
	UserID    string
	ImageID   string
	ImageData []byte
}

// NSSaveResponse contains the basic information of the image file saving
type NSSaveResponse struct {
	SaveError error `json:"error"`
}

// MakeNSStoreEndpoint upload the content file
func MakeNSStoreEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSSaveRequest)
		err := svc.Save(req.UserID, req.ImageID, req.ImageData)
		return NSSaveResponse{SaveError: err}, err
	}
}

// NSFindRequest parse the basic information for find the shared access url
type NSFindRequest struct {
	UserID  string
	ImageID string
}

// NSFindResponse return the public shared access url
type NSFindResponse struct {
	URL       string `json:"url"`
	FindError error  `json:"error"`
}

// MakeNSFindEndpoint find the public access url for a given user id and image name
func MakeNSFindEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSFindRequest)
		url, err := svc.Find(req.UserID, req.ImageID)
		return NSFindResponse{URL: url, FindError: err}, err
	}
}
