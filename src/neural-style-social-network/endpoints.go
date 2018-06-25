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

// NSDeleteFolloweebyIDRequest define the product id and the user id
type NSDeleteFolloweebyIDRequest struct {
	ProductID string
	UserID    string
}

// NSSocialErrorResponse define the general error response
type NSSocialErrorResponse struct {
	Err error
}

// NSGetSummaryByIDRequest define the product id for summary information
type NSGetSummaryByIDRequest struct {
	ProductID string
}

// NSGetSummaryByIDResponse define the summary information
type NSGetSummaryByIDResponse struct {
	Summary SocialSummary
	Err     error
}

// NSGetFolloweeProductsByUserRequest params for get following products
type NSGetFolloweeProductsByUserRequest struct {
	User string
}

// NSGetFolloweeProductsByUserResponse get the following products
type NSGetFolloweeProductsByUserResponse struct {
	Prods []FollowingProduct
	Err   error
}

//HealthResponse return the status of the health
type HealthResponse struct {
	Status bool `json:"status"`
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
		return NSSocialErrorResponse{Err: err}, err
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
		return NSSocialErrorResponse{Err: err}, err
	}
}

func makeNSDeleteFolloweeByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSDeleteFolloweebyIDRequest)
		err := svc.DeleteFolloweeByID(req.ProductID, req.UserID)
		return NSSocialErrorResponse{Err: err}, err
	}
}

func makeNSGetSummaryByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetSummaryByIDRequest)
		summary, err := svc.GetSummaryByID(req.ProductID)
		return NSGetSummaryByIDResponse{Summary: summary, Err: err}, err
	}
}

func makeNSGetFolloweeProductsByUserEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetFolloweeProductsByUserRequest)
		prods, err := svc.GetFollowingProductsByUserID(req.User)
		return NSGetFolloweeProductsByUserResponse{Prods: prods, Err: err}, err
	}
}

func makeNSGetHealthEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		status := svc.HealthCheck()
		return HealthResponse{Status: status}, nil
	}
}
