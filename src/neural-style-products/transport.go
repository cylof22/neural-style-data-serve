package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-kit/kit/log"
	mgo "gopkg.in/mgo.v2"

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

func decodeNSCacheGetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	userID := vars["usrid"]
	imageID := vars["imgid"]

	return NSCacheGetRequest{UserID: userID, ImageID: imageID}, nil
}

func encodeNSCachedGetResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	getRes := response.(NSCacheGetResponse)

	if getRes.Error != nil {
		// Todo: add error log
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", getRes.Type)
	imgSize := len(getRes.Data)
	w.Header().Set("Content-Length", strconv.FormatInt(int64(imgSize), 10))

	length, err := w.Write(getRes.Data)

	if length != len(getRes.Data) {
		return errors.New("Empty image")
	}

	return err
}

func decodeNSDeleteProductRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	return NSDeleteProductRequest{ID: id}, nil
}

func encodeNSDeleteProductResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	deleteError := response.(NSDeleteProductResponse)
	if deleteError.Err != nil {
		return deleteError.Err
	}

	return nil
}

func decodeNSUpdateProductRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	productData := UploadProduct{}
	json.NewDecoder(r.Body).Decode(&productData)

	return NSUpdateProductRequest{ID: id, ProductData: productData}, nil
}

func encodeNSUpdateProductResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	updateError := response.(NSUpdateProductResponse)
	if updateError.Err != nil {
		return updateError.Err
	}

	return nil
}

func decodeNSUpdateProductAfterTransactionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	id := vars["id"]

	updateData := NSUtil.TransactionUpdateData{}
	json.NewDecoder(r.Body).Decode(&updateData)

	return NSUpdateProductAfterTransactionRequest{ID: id, NewOwner: updateData.Owner, NewPrice: updateData.Price}, nil
}

func encodeNSUpdateProductAfterTransactionResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	updateError := response.(NSUpdateProductAfterTransactionResponse)
	if updateError.Err != nil {
		return updateError.Err
	}

	return nil
}

func decodeNSGetProductsByUserRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	userID := vars["usrid"]

	return NSGetProductsByUserRequest{User: userID}, nil
}

func encodeGetProductsByUserResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	productsRes := response.(NSGetProductsByUserResponse)
	if productsRes.Err != nil {
		return productsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(productsRes.Prods)
}

func decodeNSGetProductsByTagsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	tags := vars["tags"]

	return NSGetProductsByTagsRequest{Tags: []string{tags}}, nil
}

func encodeNSGetProductsByTargsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	productsRes := response.(NSGetProductsByTagsResponse)
	if productsRes.Err != nil {
		return productsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(productsRes.Prods)
}

func decodeNSSearchRequest(_ context.Context, r *http.Request) (interface{}, error) {
	querys := r.URL.Query()

	// add check the validation of the query string
	queryInfo := make(map[string]interface{})
	for key, val := range querys {
		queryInfo[key] = val
	}
	return NSSearchRequest{Info: queryInfo}, nil
}

func encodeNSSearchRespones(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	productsRes := response.(NSSearchResponse)
	if productsRes.Err != nil {
		return productsRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(productsRes.Prods)
}

func makeHTTPHandler(ctx context.Context, session *mgo.Session, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(NSUtil.EncodeError),
		httptransport.ServerBefore(NSUtil.ParseToken),
	}

	// Product service
	storageServiceURL := "http://" + *storageServerURL + ":" + *storageServerPort
	storageSaveURL := storageServiceURL + *storageServerSaveRouter
	storageFindURL := storageServiceURL + *storageServerFindRouter

	cacheServiceURL := "http://" + *cacheServer
	cacheGetURL := cacheServiceURL + *cacheGetRouter

	svc := NewProductSVC(*outputPath, *serverPort,
		storageSaveURL, storageFindURL, cacheGetURL, *localDev, logger, session)

	svc = NewLoggingService(log.With(logger, "component", "product"), svc)

	auth := NSUtil.AuthMiddleware(logger)

	// POST /api/upload/content
	contentUploadHandler := httptransport.NewServer(
		auth(MakeNSContentUploadEndpoint(svc)),
		decodeNSUploadContentRequest,
		encodeNSUploadContentResponse,
		options...,
	)
	r.Methods("POST").Path("/api/upload/content").Handler(NSUtil.AccessControl(contentUploadHandler))

	// POST /api/upload/style
	styleUploadHandler := httptransport.NewServer(
		auth(MakeNSStyleUploadEndpoint(svc)),
		decodeNSUploadStyleRequest,
		encodeNSUploadStyleResponse,
		options...,
	)
	r.Methods("POST").Path("/api/upload/style").Handler(NSUtil.AccessControl(styleUploadHandler))

	// POST /api/upload/styles
	stylesUploadHandler := httptransport.NewServer(
		auth(MakeNSStylesUploadEndpoint(svc)),
		decodeNSUploadStylesRequest,
		encodeNSUploadStylesResponse,
		options...,
	)
	r.Methods("POST").Path("/api/upload/styles").Handler(NSUtil.AccessControl(stylesUploadHandler))

	// GET /api/artists
	r.Methods("GET").Path("/api/artists").Handler(NSUtil.AccessControl(httptransport.NewServer(
		auth(MakeNSGetArtists(svc)),
		decodeNSGetArtistRequest,
		encodeNSGetArtistsResponse,
		options...,
	)))

	// GET /api/artists/hotest
	r.Methods("GET").Path("/api/artists/hotest").Handler(NSUtil.AccessControl(httptransport.NewServer(
		auth(MakeNSGetHotestArtists(svc)),
		decodeNSGetArtistRequest,
		encodeNSGetArtistsResponse,
		options...,
	)))

	// GET api/products
	r.Methods("GET").Path("/api/products").Handler(httptransport.NewServer(
		MakeNSGetProductsEndpoint(svc),
		decodeNSGetProductsRequest,
		encodeNSGetProductsResponse,
		options...,
	))

	// GET api/products/{userid}
	r.Methods("GET").Path("/api/products/user/{usrid}").Handler(httptransport.NewServer(
		auth(MakeNSGetProductsByUser(svc)),
		decodeNSGetProductsByUserRequest,
		encodeGetProductsByUserResponse,
		options...,
	))

	// GET api/products/{tags}
	r.Methods("GET").Path("/api/products/tags/{tags}").Handler(httptransport.NewServer(
		MakeNSGetProductsByTags(svc),
		decodeNSGetProductsByTagsRequest,
		encodeNSGetProductsByTargsResponse,
		options...,
	))

	// GET api/products/{id}
	r.Methods("GET").Path("/api/products/{id}").Handler(httptransport.NewServer(
		MakeNSGetProductByIDEndpoint(svc),
		decodeNSGetProductByIDRequest,
		encodeNSGetProductByIDResponse,
		options...,
	))

	// GET api/search
	r.Methods("GET").Path("/api/search").Handler(httptransport.NewServer(
		MakeNSSearch(svc),
		decodeNSSearchRequest,
		encodeNSSearchRespones,
		options...,
	))

	r.Methods("GET").Path("/api/v1/cache/get/{usrid}/{imgid}").Handler(
		httptransport.NewServer(
			MakeNSImageCacheGetEndpoint(svc),
			decodeNSCacheGetRequest,
			encodeNSCachedGetResponse,
			options...,
		))

	// DELETE api/products/{id}/delete
	r.Methods("DELETE").Path("/api/products/{id}/delete").Handler(httptransport.NewServer(
		auth(MakeNSDeleteProductEndpoint(svc)),
		decodeNSDeleteProductRequest,
		encodeNSDeleteProductResponse,
		options...,
	))

	// POST api/products/{id}/update
	r.Methods("POST").Path("/api/products/{id}/update").Handler(httptransport.NewServer(
		auth(MakeNSUpdateProductEndpoint(svc)),
		decodeNSUpdateProductRequest,
		encodeNSUpdateProductResponse,
		options...,
	))

	// POST /api/products/{id}/transactionupdate/
	r.Methods("POST").Path("/api/products/{id}/transactionupdate").Handler(httptransport.NewServer(
		MakeNSUpdateProductAfterTransactionEndpoint(svc),
		decodeNSUpdateProductAfterTransactionRequest,
		encodeNSUpdateProductAfterTransactionResponse,
		options...,
	))

	// add tencent yun authorization ssl file
	authFiles := http.FileServer(http.Dir("data/auth/"))
	r.PathPrefix("/.well-known/pki-validation/").Handler(http.StripPrefix("/.well-known/pki-validation/", authFiles))

	return r
}
