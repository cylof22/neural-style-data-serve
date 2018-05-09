package ProductService

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// NSUploadRequest parameters for upload file
type NSUploadRequest struct {
	ProductData Product
}

type NSStyleUploadRequest struct {
	ProductData UploadProduct
}

type NSStylesUploadRequest struct {
	ProductsData BatchProducts
}

type NSQueryRequest struct {
	QueryData QueryParams
}

// NSGetProductsResponse output the json response
type NSGetProductsResponse struct {
	Products []Product
	Err      error
}

// NSGetProductByIDRequest define the input parameter for get product by id
type NSGetProductByIDRequest struct {
	ID string
}

// NSGetProductResponse output the selected product by id
type NSGetProductResponse struct {
	Target Product
	Err    error
}

type NSUploadProductsResponse struct {
	Result string
	Err    error
}

// NSGetReviewsByIDRequest define the parameters for get reviews
type NSGetReviewsByIDRequest struct {
	ID string
}

// NSGetReviewsByIDResponse output the selected reviews
type NSGetReviewsByIDResponse struct {
	Reviews []Review
	Err     error
}

// NSGetArtistsResponse return supported artists
type NSGetArtistsResponse struct {
	Artists []Artist
	Err     error
}

// NSCacheGetRequest define request key
type NSCacheGetRequest struct {
	UserID  string
	ImageID string
}

// NSCacheGetResponse define the cached image data
type NSCacheGetResponse struct {
	Data  []byte
	Type  string
	Error error
}

type NSDeleteProductRequest struct {
	ID string
}

type NSDeleteProductResponse struct {
	Err error
}

type NSUpdateProductRequest struct {
	ID string
	ProductData UploadProduct
}

type NSUpdateProductResponse struct {
	Err error
}

// MakeNSContentUploadEndpoint upload the content file
func MakeNSContentUploadEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSUploadRequest)
		prod, err := svc.UploadContentFile(req.ProductData)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSStyleUploadEndpoint upload the style file
func MakeNSStyleUploadEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSStyleUploadRequest)
		prod, err := svc.UploadStyleFile(req.ProductData)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSStylesUploadEndpoint upload the style file
func MakeNSStylesUploadEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSStylesUploadRequest)
		res, err := svc.UploadStyleFiles(req.ProductsData)
		return NSUploadProductsResponse{Result: res, Err: err}, err
	}
}

// MakeNSGetProductsEndpoint get all the transfered file
func MakeNSGetProductsEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSQueryRequest)
		output, err := svc.GetProducts(req.QueryData)
		return NSGetProductsResponse{Products: output, Err: err}, err
	}
}

// MakeNSGetProductByIDEndpoint get the selected product by id
func MakeNSGetProductByIDEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetProductByIDRequest)
		prod, err := svc.GetProductsByID(req.ID)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSGetReviewsByIDEndpoint get the selected reviews by id
func MakeNSGetReviewsByIDEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetReviewsByIDRequest)
		reviews, err := svc.GetReviewsByProductID(req.ID)
		return NSGetReviewsByIDResponse{Reviews: reviews, Err: err}, err
	}
}

// MakeNSGetArtists generate the endpoint for get hotest artists
func MakeNSGetArtists(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		artists, err := svc.GetArtists()
		return NSGetArtistsResponse{Artists: artists, Err: err}, err
	}
}

// MakeNSGetHotestArtists generate the endpoint for getting hotest artists
func MakeNSGetHotestArtists(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		hotestArtists, err := svc.GetHotestArtists()
		return NSGetArtistsResponse{Artists: hotestArtists, Err: err}, err
	}
}

// MakeNSImageCacheGetEndpoint define the endpoint for image cache get
func MakeNSImageCacheGetEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSCacheGetRequest)
		data, mimeType, err := svc.GetImage(req.UserID, req.ImageID)
		return NSCacheGetResponse{Data: data, Type: mimeType, Error: err}, err
	}
}

// MakeNSDeleteProductEndpoint deletes one product
func MakeNSDeleteProductEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSDeleteProductRequest)
		err := svc.DeleteProduct(req.ID)
		return NSDeleteProductResponse{Err: err}, err
	}
}

// MakeNSUpdateProductEndpoint updates one product
func MakeNSUpdateProductEndpoint(svc *ProductService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSUpdateProductRequest)
		err := svc.UpdateProduct(req.ID, req.ProductData)
		return NSUpdateProductResponse{Err: err}, err
	}
}
