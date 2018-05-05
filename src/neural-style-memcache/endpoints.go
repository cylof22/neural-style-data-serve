package main

import (
	"context"
	"net/http"

	"neural-style-util"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

// NSCacheSaveRequest define the basic parameters for memcached
type NSCacheSaveRequest struct {
	Key  string
	Data []byte
}

// NSCacheSaveResponse define error information
type NSCacheSaveResponse struct {
	Error error
}

// MakeNSImageCacheSaveEndpoint define the endpoint for image cache
func MakeNSImageCacheSaveEndpoint(svc *CacheService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSCacheSaveRequest)
		err := svc.AddImage(req.Key, req.Data)
		return NSCacheSaveResponse{Error: err}, err
	}
}

// NSCacheGetRequest define request key
type NSCacheGetRequest struct {
	Key string
}

// NSCacheGetResponse define the cached image data
type NSCacheGetResponse struct {
	Data  []byte
	Type  string
	Error error
}

// MakeNSImageCacheGetEndpoint define the endpoint for image cache get
func MakeNSImageCacheGetEndpoint(svc *CacheService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSCacheGetRequest)
		data, mimeType, err := svc.GetImage(req.Key)

		return NSCacheGetResponse{Data: data, Type: mimeType, Error: err}, err
	}
}
func makeHTTPHandler(ctx context.Context, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(NSUtil.EncodeError),
	}

	var memcachedURL []string
	memcachedURL = append(memcachedURL, "localhost:11211")
	svc := NewCacheService(memcachedURL...)

	r.Methods("POST").Path("/api/v1/cache/save").Queries("key", "{key}").Handler(
		httptransport.NewServer(
			MakeNSImageCacheSaveEndpoint(svc),
			decodeNSCacheSaveRequest,
			encodeNSCacheSaveResponse,
			options...,
		))

	r.Methods("GET").Path("/api/v1/cache/get/{key}").Handler(
		httptransport.NewServer(
			MakeNSImageCacheGetEndpoint(svc),
			decodeNSCacheGetRequest,
			encodeNSCachedGetResponse,
			options...,
		))
	return r
}
