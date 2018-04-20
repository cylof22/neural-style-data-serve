package ProductService

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// NSUploadRequest parameters for upload file
type NSUploadRequest struct {
	ProductData Product
}

// NSGetProductsResponse output the json response
type NSGetProductsResponse struct {
	Products []Product
	Err      error
}

// NSGetProductByIDRequest define the input parameter for get product by id
type NSGetProductByIDRequest struct {
	ID string
}

// NSGetProductResponse output the selected product by id
type NSGetProductResponse struct {
	Target Product
	Err    error
}

// NSGetReviewsByIDRequest define the parameters for get reviews
type NSGetReviewsByIDRequest struct {
	ID string
}

// NSGetReviewsByIDResponse output the selected reviews
type NSGetReviewsByIDResponse struct {
	Reviews []Review
	Err     error
}

// NSGetArtistsResponse return supported artists
type NSGetArtistsResponse struct {
	Artists []Artist
	Err     error
}

// MakeNSContentUploadEndpoint upload the content file
func MakeNSContentUploadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSUploadRequest)
		prod, err := svc.UploadContentFile(req.ProductData)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSStyleUploadEndpoint upload the style file
func MakeNSStyleUploadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSUploadRequest)
		prod, err := svc.UploadStyleFile(req.ProductData)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSGetProductsEndpoint get all the transfered file
func MakeNSGetProductsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		output, err := svc.GetProducts()
		return NSGetProductsResponse{Products: output, Err: err}, err
	}
}

// MakeNSGetProductByIDEndpoint get the selected product by id
func MakeNSGetProductByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetProductByIDRequest)
		prod, err := svc.GetProductsByID(req.ID)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSGetReviewsByIDEndpoint get the selected reviews by id
func MakeNSGetReviewsByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetReviewsByIDRequest)
		reviews, err := svc.GetReviewsByProductID(req.ID)
		return NSGetReviewsByIDResponse{Reviews: reviews, Err: err}, err
	}
}

// MakeNSGetArtists generate the endpoint for get hotest artists
func MakeNSGetArtists(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		artists, err := svc.GetArtists()
		return NSGetArtistsResponse{Artists: artists, Err: err}, err
	}
}

// MakeNSGetHotestArtists generate the endpoint for getting hotest artists
func MakeNSGetHotestArtists(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		hotestArtists, err := svc.GetHotestArtists()
		return NSGetArtistsResponse{Artists: hotestArtists, Err: err}, err
	}
}
