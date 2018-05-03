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

	svc := NewStorageService(dbSession)

	//POST /api/storage/save/v1/{userid}/{imageid}
	r.Methods("POST").Path("/api/storage/save/v1").Queries("userid", "{userid}", "imageid", "{imageid}").Handler(
		httptransport.NewServer(
			MakeNSStoreEndpoint(svc),
			decodeNSSaveRequest,
			encodeNSSaveResponse,
			options...,
		))

	// GET /api/storage/find/v1/{userid}/{imageid}
	r.Methods("GET").Path("/api/storage/find/v1").Queries("userid", "{userid}", "imageid", "{imageid}").Handler(
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
func MakeNSStoreEndpoint(svc *StorageService) endpoint.Endpoint {
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
func MakeNSFindEndpoint(svc *StorageService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSFindRequest)
		url, err := svc.Find(req.UserID, req.ImageID)
		return NSFindResponse{URL: url, FindError: err}, err
	}
}
