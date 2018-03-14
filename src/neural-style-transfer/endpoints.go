package StyleService

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// NeuralStyleRequest parameters for the style transfer
type NeuralStyleRequest struct {
	Content    string
	Style      string
	Iterations int
}

// NeuralStylePreviewRequest parameters for the style transfer preview
type NeuralStylePreviewRequest struct {
	Content string
	Style   string
}

// NeuralStyleResponse error information for the style transfer
type NeuralStyleResponse struct {
	Err    error
	Output string
}

// Endpoints wrap the Neural Style Service
type Endpoints struct {
	NeuralStyleEndpoint        endpoint.Endpoint
	NeuralStylePreviewEndpoint endpoint.Endpoint
}

// MakeNeuralStyleEndpoint generate style transfer endpoint
func MakeNeuralStyleEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NeuralStyleRequest)
		output, err := svc.StyleTransfer(req.Content, req.Style, req.Iterations)
		return NeuralStyleResponse{Err: err, Output: output}, err
	}
}

// MakeNeuralStylePreviewEndpoint generate the style transfer preview endpoint
func MakeNeuralStylePreviewEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NeuralStylePreviewRequest)
		output, err := svc.StyleTransferPreview(req.Content, req.Style)
		return NeuralStyleResponse{Err: err, Output: output}, err
	}
}
