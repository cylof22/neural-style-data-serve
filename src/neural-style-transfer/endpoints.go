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
	FileName string
}

// NSResponse error information for the style transfer
type NSResponse struct {
	Err    error  `json:"err"`
	Output string `json:"output"`
}

// Endpoints wrap the Neural Style Service
type Endpoints struct {
	NSEndpoint              endpoint.Endpoint
	NSPreviewEndpoint       endpoint.Endpoint
	NSContentUploadEndpoint endpoint.Endpoint
	NSStyleUploadEndpoint   endpoint.Endpoint
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
		output, err := svc.UploadContentFile(req.FileName, nil)
		return NSResponse{Err: err, Output: output}, err
	}
}

// MakeNSStyleUploadEndpoint upload the style file
func MakeNSStyleUploadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSUploadRequest)
		output, err := svc.UploadStyleFile(req.FileName, nil)
		return NSResponse{Err: err, Output: output}, err
	}
}
