package StyleService

import (
	"context"
	"encoding/base64"
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

	//GET /styleTransfer/{content}/{style}/{iterations}
	r.Methods("GET").Path("/styleTransfer").Queries("content", "{content}", "style", "{style}", "iterations", "{iterations:[0-9]+}").Handler(httptransport.NewServer(
		endpoint.NeuralStyleEndpoint,
		decodeNeuralStyleRequest,
		encodeNeuralStyleResponse,
		options...,
	))

	//GET /styleTransferPreview/{content}/{style}
	r.Methods("GET").Path("/styleTransferPreview").Queries("content", "{content}", "style", "{style}").Handler(httptransport.NewServer(
		endpoint.NeuralStylePreviewEndpoint,
		decodeNeuralStylePreviewRequest,
		encodeNeuralStyleResponse,
		options...,
	))

	// output file server
	outputFiles := http.FileServer(http.Dir("data/outputs/"))
	r.PathPrefix("/outputs/").Handler(http.StripPrefix("/outputs/", outputFiles))

	// style file server
	styleFiles := http.FileServer(http.Dir("data/styles/"))
	r.PathPrefix("/styles/").Handler(http.StripPrefix("/styles", styleFiles))

	// content file server
	contentFiles := http.FileServer(http.Dir("data/contents"))
	r.PathPrefix("/contents/").Handler(http.StripPrefix("/contents/", contentFiles))
	return r
}
func decodeNeuralStylePreviewRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	contentPath, stylePath, ok := decodeNeuralStyleCommonParams(vars)
	if ok != nil {
		return nil, ok
	}

	return NeuralStylePreviewRequest{
		Content: string(contentPath),
		Style:   string(stylePath),
	}, nil
}

func decodeNeuralStyleRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	contentPath, stylePath, ok := decodeNeuralStyleCommonParams(vars)
	if ok != nil {
		return nil, ok
	}

	iterations, isOk := vars["iterations"]
	if !isOk {
		return nil, ErrBadRouting
	}

	iterationTimes, _ := strconv.Atoi(iterations)

	return NeuralStyleRequest{
		Content:    string(contentPath),
		Style:      string(stylePath),
		Iterations: iterationTimes,
	}, nil
}

func decodeNeuralStyleCommonParams(vars map[string]string) (string, string, error) {
	content, ok := vars["content"]
	if !ok {
		return "", "", ErrBadRouting
	}
	contentPath, _ := base64.StdEncoding.DecodeString(content)

	style, ok := vars["style"]
	if !ok {
		return "", "", ErrBadRouting
	}
	stylePath, _ := base64.StdEncoding.DecodeString(style)

	return string(contentPath), string(stylePath), nil
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
