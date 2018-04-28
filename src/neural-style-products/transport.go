package ProductService

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"neural-style-util"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func decodeNSUploadContentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	productData := Product{ID: "1"}
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
	productData := UploadProduct{}
	json.NewDecoder(r.Body).Decode(&productData)
	return NSStyleUploadRequest{ProductData: productData}, nil
}

func encodeNSUploadStyleResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	styleRes := response.(NSGetProductResponse)
	if styleRes.Err != nil {
		return styleRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(styleRes.Target)
}

func decodeNSUploadStylesRequest(_ context.Context, r *http.Request) (interface{}, error) {
	productsData := BatchProducts{}
	json.NewDecoder(r.Body).Decode(&productsData)
	return NSStylesUploadRequest{ProductsData: productsData}, nil
}

func encodeNSUploadStylesResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	uploadRes := response.(NSUploadProductsResponse)
	if uploadRes.Err != nil {
		return uploadRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(uploadRes.Result)
}

func decodeNSGetProductsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	queryData, _ := url.ParseQuery(r.URL.RawQuery)
	queryBytes, _ := json.Marshal(queryData)

	var params QueryParams
	json.Unmarshal(queryBytes, &params)
	return NSQueryRequest{QueryData:params}, nil
}

func encodeNSGetProductsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	productsRes := response.(NSGetProductsResponse)
	if productsRes.Err != nil {
		return productsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(productsRes.Products)
}

func decodeNSGetArtistRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeNSGetArtistsResponse(ctx context.Context, w http.ResponseWriter, res interface{}) error {
	artistsRes := res.(NSGetArtistsResponse)
	if artistsRes.Err != nil {
		return artistsRes.Err
	}

	w.Header().Set("content-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(artistsRes.Artists)
}

func decodeNSGetProductByIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	return NSGetProductByIDRequest{ID: id}, nil
}

func encodeNSGetProductByIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
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

// MakeHTTPHandler generate the http handler for the style service handler
func MakeHTTPHandler(ctx context.Context, r *mux.Router, svc *ProductService, options ...httptransport.ServerOption) *mux.Router {
	// POST /api/upload/content
	contentUploadHandler := httptransport.NewServer(
		MakeNSContentUploadEndpoint(svc),
		decodeNSUploadContentRequest,
		encodeNSUploadContentResponse,
		options...,
	)
	r.Methods("POST").Path("/api/upload/content").Handler(NSUtil.AuthMiddleware(NSUtil.AccessControl(contentUploadHandler)))

	// POST /api/upload/style
	styleUploadHandler := httptransport.NewServer(
		MakeNSStyleUploadEndpoint(svc),
		decodeNSUploadStyleRequest,
		encodeNSUploadStyleResponse,
		options...,
	)
	r.Methods("POST").Path("/api/upload/style").Handler(NSUtil.AuthMiddleware(NSUtil.AccessControl(styleUploadHandler)))

	// POST /api/upload/styles
	stylesUploadHandler := httptransport.NewServer(
		MakeNSStylesUploadEndpoint(svc),
		decodeNSUploadStylesRequest,
		encodeNSUploadStylesResponse,
		options...,
	)
	r.Methods("POST").Path("/api/upload/styles").Handler(NSUtil.AuthMiddleware(NSUtil.AccessControl(stylesUploadHandler)))

	// GET /api/artists
	r.Methods("GET").Path("/api/artists").Handler(NSUtil.AuthMiddleware(NSUtil.AccessControl(httptransport.NewServer(
		MakeNSGetArtists(svc),
		decodeNSGetArtistRequest,
		encodeNSGetArtistsResponse,
		options...,
	))))

	// GET /api/artists/hotest
	r.Methods("GET").Path("/api/artists/hotest").Handler(NSUtil.AuthMiddleware(NSUtil.AccessControl(httptransport.NewServer(
		MakeNSGetHotestArtists(svc),
		decodeNSGetArtistRequest,
		encodeNSGetArtistsResponse,
		options...,
	))))

	// GET api/products
	r.Methods("GET").Path("/api/products").Handler(NSUtil.UsernameMiddleware(httptransport.NewServer(
		MakeNSGetProductsEndpoint(svc),
		decodeNSGetProductsRequest,
		encodeNSGetProductsResponse,
		options...,
	)))

	// GET api/products/{id}
	r.Methods("GET").Path("/api/products/{id}").Handler(httptransport.NewServer(
		MakeNSGetProductByIDEndpoint(svc),
		decodeNSGetProductByIDRequest,
		encodeNSGetProductByIDResponse,
		options...,
	))

	// GET api/products/{id}/reviews
	r.Methods("GET").Path("/api/products/{id}/reviews").Handler(NSUtil.AuthMiddleware(httptransport.NewServer(
		MakeNSGetReviewsByIDEndpoint(svc),
		decodeNSGetReviewsByIDRequest,
		encodeNSGetReviewsByIDResponse,
		options...,
	)))

	// Todo: Web Service maybe need a seperate server
	// output file server
	outputFiles := http.FileServer(http.Dir("data/outputs/"))
	r.PathPrefix("/outputs/").Handler(http.StripPrefix("/outputs/", outputFiles))

	// style file server
	styleFiles := http.FileServer(http.Dir("data/styles/"))
	r.PathPrefix("/styles/").Handler(http.StripPrefix("/styles", styleFiles))

	// content file server
	contentFiles := http.FileServer(http.Dir("data/contents"))
	r.PathPrefix("/contents/").Handler(http.StripPrefix("/contents/", contentFiles))

	// template file
	resourceFile := http.FileServer(http.Dir("dist"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", resourceFile))

	r.Path("/").Handler(resourceFile)

	return r
}
