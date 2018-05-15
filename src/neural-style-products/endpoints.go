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

// NSDeleteProductRequest define the id of the deleted product
type NSDeleteProductRequest struct {
	ID string
}

// NSDeleteProductResponse only returns the error information for delete from DB
type NSDeleteProductResponse struct {
	Err error
}

// NSUpdateProductRequest define the id of the updated product and its updated information
type NSUpdateProductRequest struct {
	ID          string
	ProductData UploadProduct
}

// NSUpdateProductResponse only returns the error information for updating a product
type NSUpdateProductResponse struct {
	Err error
}

// NSGetProductsByUserRequest define the use who want to get its own products
type NSGetProductsByUserRequest struct {
	User string
}

// NSGetProductsByUserResponse return the error information and the use's products
type NSGetProductsByUserResponse struct {
	Prods []Product
	Err   error
}

// NSGetProductsByTagsRequest define the tags for getting the products
type NSGetProductsByTagsRequest struct {
	Tags []string
}

// NSGetProductsByTagsResponse return the products which contains the tags
type NSGetProductsByTagsResponse struct {
	Prods []Product
	Err   error
}

// NSSearchRequest define the search info for getting the products
type NSSearchRequest struct {
	Info map[string]interface{}
}

// NSSearchResponse return the searched products and the error information
type NSSearchResponse struct {
	Prods []Product
	Err   error
}

// MakeNSContentUploadEndpoint upload the content file
func MakeNSContentUploadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSUploadRequest)
		prod, err := svc.UploadContentFile(req.ProductData)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSStyleUploadEndpoint upload the style file
func MakeNSStyleUploadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSStyleUploadRequest)
		prod, err := svc.UploadStyleFile(req.ProductData)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSStylesUploadEndpoint upload the style file
func MakeNSStylesUploadEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSStylesUploadRequest)
		res, err := svc.UploadStyleFiles(req.ProductsData)
		return NSUploadProductsResponse{Result: res, Err: err}, err
	}
}

// MakeNSGetProductsEndpoint get all the transfered file
func MakeNSGetProductsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		output, err := svc.GetProducts()
		return NSGetProductsResponse{Products: output, Err: err}, err
	}
}

// MakeNSGetProductByIDEndpoint get the selected product by id
func MakeNSGetProductByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetProductByIDRequest)
		prod, err := svc.GetProductsByID(req.ID)
		return NSGetProductResponse{Target: prod, Err: err}, err
	}
}

// MakeNSGetReviewsByIDEndpoint get the selected reviews by id
func MakeNSGetReviewsByIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetReviewsByIDRequest)
		reviews, err := svc.GetReviewsByProductID(req.ID)
		return NSGetReviewsByIDResponse{Reviews: reviews, Err: err}, err
	}
}

// MakeNSGetArtists generate the endpoint for get hotest artists
func MakeNSGetArtists(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		artists, err := svc.GetArtists()
		return NSGetArtistsResponse{Artists: artists, Err: err}, err
	}
}

// MakeNSGetHotestArtists generate the endpoint for getting hotest artists
func MakeNSGetHotestArtists(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		hotestArtists, err := svc.GetHotestArtists()
		return NSGetArtistsResponse{Artists: hotestArtists, Err: err}, err
	}
}

// MakeNSImageCacheGetEndpoint define the endpoint for image cache get
func MakeNSImageCacheGetEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSCacheGetRequest)
		data, mimeType, err := svc.GetImage(req.UserID, req.ImageID)
		return NSCacheGetResponse{Data: data, Type: mimeType, Error: err}, err
	}
}

// MakeNSDeleteProductEndpoint deletes one product
func MakeNSDeleteProductEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSDeleteProductRequest)
		err := svc.DeleteProduct(req.ID)
		return NSDeleteProductResponse{Err: err}, err
	}
}

// MakeNSUpdateProductEndpoint updates one product
func MakeNSUpdateProductEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSUpdateProductRequest)
		err := svc.UpdateProduct(req.ID, req.ProductData)
		return NSUpdateProductResponse{Err: err}, err
	}
}

// MakeNSGetProductsByUser get all products owned users
func MakeNSGetProductsByUser(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetProductsByUserRequest)
		prods, err := svc.GetProductsByUser(req.User)
		return NSGetProductsByUserResponse{Prods: prods}, err
	}
}

// MakeNSGetProductsByTags get all products related to the tags
func MakeNSGetProductsByTags(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetProductsByTagsRequest)
		prods, err := svc.GetProductsByTags(req.Tags)
		return NSGetProductsByTagsResponse{Prods: prods}, err
	}
}

// MakeNSSearch return the searched products by following the keywords
func MakeNSSearch(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSSearchRequest)
		prods, err := svc.Search(req.Info)
		return NSSearchResponse{Prods: prods, Err: err}, err
	}
}
