package StyleService

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

var (
	// ErrBadRouting define the default routing error information
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// MakeHTTPHandler generate the http handler for the style service handler
func MakeHTTPHandler(ctx context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	//GET /styleTransfer/{content}/{style}/{output}/{iterations}
	r.Methods("GET").Path("/styleTransfer").Queries("content", "{content}", "style", "{style}", "output", "{output}", "iterations", "{iterations:[0-9]+}").Handler(httptransport.NewServer(
		endpoint.NeuralStyleEndpoint,
		decodeNeuralStyleRequest,
		encodeNeuralStyleResponse,
		options...,
	))

	return r
}

func decodeNeuralStyleRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	content, ok := vars["content"]
	if !ok {
		return nil, ErrBadRouting
	}

	style, ok := vars["style"]
	if !ok {
		return nil, ErrBadRouting
	}

	output, ok := vars["output"]
	if !ok {
		return nil, ErrBadRouting
	}

	iterations, ok := vars["iterations"]
	if !ok {
		return nil, ErrBadRouting
	}

	iterationTimes, _ := strconv.Atoi(iterations)

	return NeuralStyleRequest{
		Content:    content,
		Style:      style,
		Output:     output,
		Iterations: iterationTimes,
	}, nil
}

type errorer interface {
	error() error
}

func encodeNeuralStyleResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		// Not a gokit transport error, but a business logic error
		// provide those as HTTP errors
		encodeError(ctx, e.error(), w)
		return nil
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(response)
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	if err != nil {
		panic("encodeError with nil error")
	}

	w.Header().Set("context-type", "application/json,charset=utf8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
