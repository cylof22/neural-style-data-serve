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

// NeuralStyleResponse error information for the style transfer
type NeuralStyleResponse struct {
	Err error `json:"err,omitempty"`
}

// Endpoints wrap the Neural Style Service
type Endpoints struct {
	NeuralStyleEndpoint endpoint.Endpoint
}

// MakeNeuralStyleEndpoint generate style transfer endpoint
func MakeNeuralStyleEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NeuralStyleRequest)
		err := svc.StyleTransfer(req.Content, req.Style, req.Output, req.Iterations)
		return NeuralStyleResponse{err}, err
	}
}
