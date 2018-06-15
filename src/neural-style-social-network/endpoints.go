package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

// NSGetReviewsByIDRequest define the parameters for get reviews
type NSGetReviewsByIDRequest struct {
	ID string
}

// NSGetReviewsByIDResponse output the selected reviews
type NSGetReviewsByIDResponse struct {
	Reviews []Review
	Err     error
}

// NSAddReviewByIDRequest define the product id and review data
type NSAddReviewByIDRequest struct {
	ID   string
	Data Review
}

// NSAddReviewByIDResponse output the internal error information
type NSAddReviewByIDResponse struct {
	Err error
}

// NSGetFolloweesByIDRequest define the parameters for get reviews
type NSGetFolloweesByIDRequest struct {
	ID string
}

// NSGetFolloweesByIDResponse output the selected reviews
type NSGetFolloweesByIDResponse struct {
	Followees []Followee
	Err       error
}

// NSAddFolloweeByIDRequest define the product id and review data
type NSAddFolloweeByIDRequest struct {
	ID   string
	Data Followee
}

// NSAddFolloweeByIDResponse output the internal error information
type NSAddFolloweeByIDResponse struct {
	Err error
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("context-type", "application/json,charset=utf8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func makeNSGetReviewsByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetReviewsByIDRequest)
		reviews, err := svc.GetReviewsByProductID(req.ID)
		return NSGetReviewsByIDResponse{Reviews: reviews, Err: err}, err
	}
}

func makeNSAddReviewByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAddReviewByIDRequest)
		err := svc.AddReviewByProductID(req.Data)
		return NSAddReviewByIDResponse{Err: err}, err
	}
}

func makeNSGetFolloweesByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetFolloweesByIDRequest)
		follows, err := svc.GetFolloweesByProductID(req.ID)
		return NSGetFolloweesByIDResponse{Followees: follows, Err: err}, err
	}
}

func makeNSAddFolloweebyIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAddFolloweeByIDRequest)
		err := svc.AddFolloweesByProductID(req.Data)
		return NSAddFolloweeByIDResponse{Err: err}, err
	}
}
