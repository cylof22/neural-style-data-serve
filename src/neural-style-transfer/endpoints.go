package StyleService

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// NeuralStyleRequest parameters for the style transfer
type NeuralStyleRequest struct {
	Content    string
	Style      string
	Output     string
	Iterations int
}

// NeuralStylePreviewRequest parameters for the style transfer preview
type NeuralStylePreviewRequest struct {
	Content string
	Style   string
	Output  string
}

// NeuralStyleResponse error information for the style transfer
type NeuralStyleResponse struct {
	Err error `json:"err,omitempty"`
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
		err := svc.StyleTransfer(req.Content, req.Style, req.Output, req.Iterations)
		return NeuralStyleResponse{err}, err
	}
}

func MakeNeuralStylePreviewEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NeuralStylePreviewRequest)
		err := svc.StyleTransferPreview(req.Content, req.Style, req.Output)
		return NeuralStyleResponse{err}, err
	}
}
