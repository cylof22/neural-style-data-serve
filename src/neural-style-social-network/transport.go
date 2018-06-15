package main

import (
	"context"
	"encoding/json"
	"net/http"

	"neural-style-util"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
)

func decodeNSGetReviewsByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	return NSGetReviewsByIDRequest{ID: id}, nil
}

func encodeNSGetReviewsByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	reviewsRes := response.(NSGetReviewsByIDResponse)
	if reviewsRes.Err != nil {
		return reviewsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(reviewsRes.Reviews)
}

func decodeAddReviewByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	review := Review{}
	err := json.NewDecoder(r.Body).Decode(&review)

	return NSAddReviewByIDRequest{ID: id, Data: review}, err
}

func encodeAddReviewByIDResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	reviewsRes := response.(NSAddReviewByIDResponse)
	if reviewsRes.Err != nil {
		return reviewsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(reviewsRes.Err)
}

func decodeNSGetFolloweesByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	return NSGetReviewsByIDRequest{ID: id}, nil
}

func encodeNSGetFolloweesByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	reviewsRes := response.(NSGetReviewsByIDResponse)
	if reviewsRes.Err != nil {
		return reviewsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(reviewsRes.Reviews)
}

func decodeAddFolloweeByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	followee := Followee{}
	err := json.NewDecoder(r.Body).Decode(&followee)

	return NSAddFolloweeByIDRequest{ID: id, Data: followee}, err
}

func encodeAddFolloweeByIDResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	followsRes := response.(NSAddFolloweeByIDResponse)
	if followsRes.Err != nil {
		return followsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(followsRes.Err)
}

func makeHTTPHandler(context context.Context, session *mgo.Session, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(NSUtil.ParseToken),
	}

	svc := newSocialSVC(logger, session)
	svc = newLoggingService(logger, svc)

	// GET api/social/v1/{id}/reviews
	r.Methods("GET").Path("/api/social/v1/{id}/reviews").Handler(httptransport.NewServer(
		makeNSGetReviewsByIDEndpoint(svc),
		decodeNSGetReviewsByIDRequest,
		encodeNSGetReviewsByIDResponse,
		options...,
	))

	// POST api/social/v1/{id}/reviews/add
	r.Methods("POST").Path("/api/social/v1/{id}/reviews/add").Handler(httptransport.NewServer(
		makeNSAddReviewByIDEndpoint(svc),
		decodeAddReviewByIDRequest,
		encodeAddReviewByIDResponse,
		options...,
	))

	// Get api/social/v1/{id}/followees
	r.Methods("GET").Path("/api/social/v1/{id}/followees").Handler(httptransport.NewServer(
		makeNSGetFolloweesByIDEndpoint(svc),
		decodeNSGetReviewsByIDRequest,
		encodeNSGetFolloweesByIDResponse,
		options...,
	))

	// POST api/social/v1/{id}/followees/add
	r.Methods("POST").Path("/api/social/v1/{id}/followees").Handler(httptransport.NewServer(
		makeNSAddFolloweebyIDEndpoint(svc),
		decodeAddFolloweeByIDRequest,
		encodeAddFolloweeByIDResponse,
		options...,
	))
	return r
}
