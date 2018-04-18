package StyleService

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// NSRequest parameters for the style transfer
type NSRequest struct {
	Content    string
	Style      string
	Iterations int
}

// NSPreviewRequest parameters for the style transfer preview
type NSPreviewRequest struct {
	Content string
	Style   string
}

// NSUploadRequest parameters for upload file
type NSUploadRequest struct {
	ProductData Product
}

// NSResponse error information for the style transfer
type NSResponse struct {
	Err    error  `json:"err"`
	Output string `json:"output"`
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

// NSAuthenticationRequest parameters for register and login
type NSAuthenticationRequest struct {
	UserData UserInfo
}

// NSRegisterResponse returns register result
type NSRegisterResponse struct {
	Result string `json:"result"`
	Err    error  `json:"err"`
}

// NSLoginResponse returns token
type NSLoginResponse struct {
	Target UserToken
	Err    error
}

// NSGetArtistsResponse return supported artists
type NSGetArtistsResponse struct {
	Artists []Artist
	Err     error
}

// Endpoints wrap the Neural Style Service
type Endpoints struct {
	NSEndpoint                 endpoint.Endpoint
	NSPreviewEndpoint          endpoint.Endpoint
	NSContentUploadEndpoint    endpoint.Endpoint
	NSStyleUploadEndpoint      endpoint.Endpoint
	NSGetProductsEndpoint      endpoint.Endpoint
	NSGetProductsByIDEndpoint  endpoint.Endpoint
	NSGetReviewsByIDEndpoint   endpoint.Endpoint
	NSRegisterEndpoint         endpoint.Endpoint
	NSLoginEndpoint            endpoint.Endpoint
	NSGetArtistsEndpoint       endpoint.Endpoint
	NSGetHotestArtistsEndpoint endpoint.Endpoint
}

// MakeNSEndpoint generate style transfer endpoint
func MakeNSEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSRequest)
		output, err := svc.StyleTransfer(req.Content, req.Style, req.Iterations)
		return NSResponse{Err: err, Output: output}, err
	}
}

// MakeNSPreviewEndpoint generate the style transfer preview endpoint
func MakeNSPreviewEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSPreviewRequest)
		output, err := svc.StyleTransferPreview(req.Content, req.Style)
		return NSResponse{Err: err, Output: output}, err
	}
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

// MakeNSRegisterEndpoint generate the endpoint for new user register
func MakeNSRegisterEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAuthenticationRequest)
		res, err := svc.Register(req.UserData)
		return NSRegisterResponse{Result: res, Err: err}, err
	}
}

// MakeNSLoginEndpoint generate the endpoint for user's login
func MakeNSLoginEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAuthenticationRequest)
		token, err := svc.Login(req.UserData)
		return NSLoginResponse{Target: token, Err: err}, err
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
