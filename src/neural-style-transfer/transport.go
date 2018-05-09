package StyleService

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/go-kit/kit/endpoint"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

var (
	// ErrBadRouting define the default routing error information
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

// templ represents a single template
type templateHandler struct {
	once     sync.Once
	filename string
	templ    *template.Template
}

// ServeHTTP handles the HTTP request.
func (t *templateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.once.Do(func() {
		t.templ = template.Must(template.ParseFiles(filepath.Join("dist",
			t.filename)))
	})
	t.templ.Execute(w, r)
}

// MakeHTTPHandler generate the http handler for the style service handler
func MakeHTTPHandler(ctx context.Context, r *mux.Router, auth endpoint.Middleware, svc *NeuralTransferService, options ...httptransport.ServerOption) *mux.Router {
	//GET /styleTransfer/{content}/{style}/{iterations}
	r.Methods("GET").Path("/styleTransfer").Queries("content", "{content}", "style", "{style}", "iterations", "{iterations:[0-9]+}").Handler(httptransport.NewServer(
		auth(MakeNSEndpoint(svc)),
		decodeNSRequest,
		encodeNSResponse,
		options...,
	))

	//GET /styleTransferPreview/{content}/{style}
	r.Methods("GET").Path("/styleTransferPreview").Queries("content", "{content}", "style", "{style}").Handler(httptransport.NewServer(
		auth(MakeNSPreviewEndpoint(svc)),
		decodeNSPreviewRequest,
		encodeNSResponse,
		options...,
	))

	return r
}

func decodeNSPreviewRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	contentPath, stylePath, ok := decodeNeuralStyleCommonParams(vars)
	if ok != nil {
		return nil, ok
	}

	return NSPreviewRequest{
		Content: string(contentPath),
		Style:   string(stylePath),
	}, nil
}

func decodeNSRequest(_ context.Context, r *http.Request) (interface{}, error) {
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

	return NSRequest{
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
	contentPath, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", "", err
	}

	style, ok := vars["style"]
	if !ok {
		return "", "", ErrBadRouting
	}
	stylePath, err := base64.StdEncoding.DecodeString(style)
	if err != nil {
		return "", "", err
	}

	return string(contentPath), string(stylePath), nil
}

type errorer interface {
	error() error
}

func encodeNSResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
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
	/* 	if err != nil {
		panic("encodeError with nil error")
	} */

	w.Header().Set("context-type", "application/json,charset=utf8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
