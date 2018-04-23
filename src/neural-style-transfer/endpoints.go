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

// NSResponse error information for the style transfer
type NSResponse struct {
	Err    error  `json:"err"`
	Output string `json:"output"`
}

// MakeNSEndpoint generate style transfer endpoint
func MakeNSEndpoint(svc *NeuralTransferService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSRequest)
		output, err := svc.StyleTransfer(req.Content, req.Style, req.Iterations)
		return NSResponse{Err: err, Output: output}, err
	}
}

// MakeNSPreviewEndpoint generate the style transfer preview endpoint
func MakeNSPreviewEndpoint(svc *NeuralTransferService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSPreviewRequest)
		output, err := svc.StyleTransferPreview(req.Content, req.Style)
		return NSResponse{Err: err, Output: output}, err
	}
}
