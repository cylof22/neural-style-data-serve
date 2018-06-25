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

func decodeNSGetFolloweesByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	return NSGetFolloweesByIDRequest{ID: id}, nil
}

func encodeNSGetFolloweesByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	followeesRes := response.(NSGetFolloweesByIDResponse)
	if followeesRes.Err != nil {
		return followeesRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(followeesRes.Followees)
}

func decodeAddFolloweeByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	followee := Followee{}
	err := json.NewDecoder(r.Body).Decode(&followee)
	return NSAddFolloweeByIDRequest{ID: id, Data: followee}, err
}

func encodeSocialResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	followsRes := response.(NSSocialErrorResponse)
	if followsRes.Err != nil {
		return followsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(followsRes.Err)
}

func decodeDeleteFolloweeByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	// Check the exists of the variables
	productID := vars["productid"]
	userID := vars["userid"]

	return NSDeleteFolloweebyIDRequest{ProductID: productID, UserID: userID}, nil
}

func decodeGetSummaryByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	productID := vars["productid"]

	return NSGetSummaryByIDRequest{ProductID: productID}, nil
}

func encodeGetSummaryByIDResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	summaryResponse := response.(NSGetSummaryByIDResponse)
	if summaryResponse.Err != nil {
		return summaryResponse.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(summaryResponse.Summary)
}

func decodeGetFolloweeProductsByUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	user := vars["user"]

	return NSGetFolloweeProductsByUserRequest{User: user}, nil
}

func encodeGetFolloweeProductsByUserResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	followeeResponse := response.(NSGetFolloweeProductsByUserResponse)

	if followeeResponse.Err != nil {
		return followeeResponse.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(followeeResponse.Prods)
}

func decodeGetNSHealthRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
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
		encodeSocialResponse,
		options...,
	))

	// Get api/social/v1/{id}/followees
	r.Methods("GET").Path("/api/social/v1/{id}/followees").Handler(httptransport.NewServer(
		makeNSGetFolloweesByIDEndpoint(svc),
		decodeNSGetFolloweesByIDRequest,
		encodeNSGetFolloweesByIDResponse,
		options...,
	))

	// POST api/social/v1/{id}/followees/add
	r.Methods("POST").Path("/api/social/v1/{id}/followees/add").Handler(httptransport.NewServer(
		makeNSAddFolloweebyIDEndpoint(svc),
		decodeAddFolloweeByIDRequest,
		encodeSocialResponse,
		options...,
	))

	// DELETE api/social/v1/{productid}/{userid}/followees/delete
	r.Methods("DELETE").Path("/api/social/v1/{productid}/{userid}/followees/delete").Handler(httptransport.NewServer(
		makeNSDeleteFolloweeByIDEndpoint(svc),
		decodeDeleteFolloweeByIDRequest,
		encodeSocialResponse,
		options...,
	))

	// GET api/social/v1/{id}/summary
	r.Methods("GET").Path("/api/social/v1/{productid}/summary").Handler(httptransport.NewServer(
		makeNSGetSummaryByIDEndpoint(svc),
		decodeGetSummaryByIDRequest,
		encodeGetSummaryByIDResponse,
		options...,
	))

	// GET api/social/v1/{user}/followees/summary
	r.Methods("GET").Path("/api/social/v1/{user}/followees/products").Handler(httptransport.NewServer(
		makeNSGetFolloweeProductsByUserEndpoint(svc),
		decodeGetFolloweeProductsByUserRequest,
		encodeGetFolloweeProductsByUserResponse,
		options...,
	))

	//GET /health
	r.Methods("GET").Path("/health").Handler(httptransport.NewServer(
		makeNSGetHealthEndpoint(svc),
		decodeGetNSHealthRequest,
		encodeResponse,
		options...,
	))
	return r
}
