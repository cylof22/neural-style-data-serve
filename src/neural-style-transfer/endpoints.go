package StyleService

import (
	"context"
	"mime/multipart"

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
	FileName string
	ImgFile  multipart.File
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

// NSGetProductByIDResponse output the selected product by id
type NSGetProductByIDResponse struct {
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

// Endpoints wrap the Neural Style Service
type Endpoints struct {
	NSEndpoint                endpoint.Endpoint
	NSPreviewEndpoint         endpoint.Endpoint
	NSContentUploadEndpoint   endpoint.Endpoint
	NSStyleUploadEndpoint     endpoint.Endpoint
	NSGetProductsEndpoint     endpoint.Endpoint
	NSGetProductsByIDEndpoint endpoint.Endpoint
	NSGetReviewsByIDEndpoint  endpoint.Endpoint
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
		output, err := svc.UploadContentFile(req.FileName, req.ImgFile)
		return NSResponse{Err: err, Output: output}, err
	}
}

// MakeNSStyleUploadEndpoint upload the style file
func MakeNSStyleUploadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSUploadRequest)
		output, err := svc.UploadStyleFile(req.FileName, req.ImgFile)
		return NSResponse{Err: err, Output: output}, err
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
		return NSGetProductByIDResponse{Target: prod, Err: err}, err
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
