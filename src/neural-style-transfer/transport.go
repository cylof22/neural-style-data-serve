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

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

var (
	// ErrBadRouting define the default routing error information
	ErrBadRouting = errors.New("inconsistent mapping between route and handler (programmer error)")
)

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")

		h.ServeHTTP(w, r)
	})
}

func webServerControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Content-Type", "text/html; text/javascript; text/css; charset=utf-8")

		h.ServeHTTP(w, r)
	})
}

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
func MakeHTTPHandler(ctx context.Context, endpoint Endpoints, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
	}

	//GET /styleTransfer/{content}/{style}/{iterations}
	r.Methods("GET").Path("/styleTransfer").Queries("content", "{content}", "style", "{style}", "iterations", "{iterations:[0-9]+}").Handler(httptransport.NewServer(
		endpoint.NSEndpoint,
		decodeNSRequest,
		encodeNSResponse,
		options...,
	))

	//GET /styleTransferPreview/{content}/{style}
	r.Methods("GET").Path("/styleTransferPreview").Queries("content", "{content}", "style", "{style}").Handler(httptransport.NewServer(
		endpoint.NSPreviewEndpoint,
		decodeNSPreviewRequest,
		encodeNSResponse,
		options...,
	))

	// POST /styleTransfer/content
	contentUploadHandler := httptransport.NewServer(
		endpoint.NSContentUploadEndpoint,
		decodeNSUploadContentRequest,
		encodeNSUploadContentResponse,
		options...,
	)
	r.Methods("POST").Path("/api/upload/content").Handler(accessControl(contentUploadHandler))

	// POST /styleTransfer/style
	styleUploadHandler := httptransport.NewServer(
		endpoint.NSStyleUploadEndpoint,
		decodeNSUploadStyleRequest,
		encodeNSUploadStyleResponse,
		options...,
	)
	r.Methods("POST").Path("/api/upload/style").Handler(accessControl(styleUploadHandler))

	// GET api/products
	r.Methods("GET").Path("/api/products").Handler(httptransport.NewServer(
		endpoint.NSGetProductsEndpoint,
		decodeNSGetProductsRequest,
		encodeNSGetProductsResponse,
		options...,
	))

	// GET api/products/{id}
	r.Methods("GET").Path("/api/products/{id}").Handler(httptransport.NewServer(
		endpoint.NSGetProductsByIDEndpoint,
		decodeNSGetProductByIDRequest,
		encodeNSGetProductByIdResponse,
		options...,
	))

	// GET api/products/{id}/reviews
	r.Methods("GET").Path("/api/products/{id}/reviews").Handler(httptransport.NewServer(
		endpoint.NSGetReviewsByIDEndpoint,
		decodeNSGetReviewsByIDRequest,
		encodeNSGetReviewsByIDResponse,
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

	r.Path("/").Handler(webServerControl(&templateHandler{filename: "index.html"}))

	// template file
	resourceFile := http.FileServer(http.Dir("dist"))
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", resourceFile))

	// js file
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", resourceFile))

	r.Path("/").Handler(webServerControl(http.FileServer(http.Dir("templates"))))

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

func decodeNSUploadContentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	productData := Product{ID:"1"}
	json.NewDecoder(r.Body).Decode(&productData)
	return NSUploadRequest{ProductData: productData}, nil
}

func encodeNSUploadContentResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	contentRes := response.(NSGetProductResponse)
	if contentRes.Err != nil {
		return contentRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(contentRes.Target)
}

func decodeNSUploadStyleRequest(_ context.Context, r *http.Request) (interface{}, error) {
	productData := Product{ID:"1"}
	json.NewDecoder(r.Body).Decode(&productData)
	return NSUploadRequest{ProductData: productData}, nil
}

func encodeNSUploadStyleResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	styleRes := response.(NSGetProductResponse)
	if styleRes.Err != nil {
		return styleRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(styleRes.Target)
}

func decodeNSGetProductsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeNSGetProductsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	productsRes := response.(NSGetProductsResponse)
	if productsRes.Err != nil {
		return productsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(productsRes.Products)
}

func decodeNSGetProductByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	return NSGetProductByIDRequest{ID: id}, nil
}

func encodeNSGetProductByIdResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	productRes := response.(NSGetProductResponse)
	if productRes.Err != nil {
		return productRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(productRes.Target)
}

func decodeNSGetReviewsByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	return NSGetReviewsByIDRequest{ID: id}, nil
}

func encodeNSGetReviewsByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	reviewsRes := response.(NSGetReviewsByIDResponse)
	if reviewsRes.Err != nil {
		return reviewsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(reviewsRes.Reviews)
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
	if err != nil {
		panic("encodeError with nil error")
	}

	w.Header().Set("context-type", "application/json,charset=utf8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
