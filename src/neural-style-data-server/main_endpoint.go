package main

import (
	"context"
	"net/http"
	"time"

	"neural-style-util"

	"neural-style-transfer"

	"github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
)

const (
	socialServiceName   = "social"
	socialServiceTag    = "v1"
	userServiceName     = "users"
	userServiceTag      = "v1"
	productsServiceName = "products"
	productsServiceTag  = "v1"
	orderServiceName    = "orders"
	orderServiceTag     = "v1"
)

func makeHTTPHandler(ctx context.Context, client consulsd.Client, dbSession *mgo.Session, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(NSUtil.EncodeError),
		httptransport.ServerBefore(NSUtil.ParseToken),
	}

	authMiddleware := NSUtil.AuthMiddleware(logger)
	// Style Service
	styleTransferService := StyleService.NewNeuralTransferSVC(*networkPath, *previewNetworkPath,
		*outputPath, *serverURL, *serverPort)
	r = StyleService.MakeHTTPHandler(ctx, r, authMiddleware, styleTransferService, options...)

	// images for explaining return
	returnFiles := http.FileServer(http.Dir("data/returns"))
	r.PathPrefix("/returns/").Handler(http.StripPrefix("/returns/", returnFiles))

	// portraits file server
	portraitsFiles := http.FileServer(http.Dir("data/portraits"))
	r.PathPrefix("/portraits/").Handler(http.StripPrefix("/portraits/", portraitsFiles))

	// output file server
	outputFiles := http.FileServer(http.Dir("data/outputs/"))
	r.PathPrefix("/outputs/").Handler(http.StripPrefix("/outputs/", outputFiles))

	// style file server
	styleFiles := http.FileServer(http.Dir("data/styles/"))
	r.PathPrefix("/styles/").Handler(http.StripPrefix("/styles", styleFiles))

	// artist masterpiece server
	masterFiles := http.FileServer(http.Dir("data/masters/"))
	r.PathPrefix("/masters/").Handler(http.StripPrefix("/masters/", masterFiles))

	// content file server
	contentFiles := http.FileServer(http.Dir("data/contents"))
	r.PathPrefix("/contents/").Handler(http.StripPrefix("/contents/", contentFiles))

	// template file
	resourceFile := http.FileServer(http.Dir("dist"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", resourceFile))

	r.Path("/").Handler(resourceFile)

	duration := 500 * time.Millisecond

	// Add API gateway for Order service
	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "GET",
		"/api/v1/transactionorders", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "GET",
		"/api/v1/orders/{username}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "GET",
		"/api/v1/sellings/{username}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "GET",
		"/api/v1/order", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "POST",
		"/api/v1/order/create", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "GET",
		"/api/v1/orders/{id}/delete", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "POST",
		"/api/v1/orders/{id}/buy", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "POST",
		"/api/v1/orders/{chainId}/chainconfirm", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "POST",
		"/api/v1/orders/{id}/productship", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "GET",
		"/api/v1/orders/{id}/confirm", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "POST",
		"/api/v1/orders/{id}/askreturn", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "POST",
		"/api/v1/orders/{id}/returnship", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "GET",
		"/api/v1/orders/{id}/returnagreed", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "GET",
		"/api/v1/orders/{id}/returnconfirmed", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, orderServiceName, orderServiceTag, "POST",
		"/api/v1/orders/{chainId}/chaincancel", duration, 3)

	// Add API gateway for Social Service
	r = NSUtil.RegisterSDService(ctx, r, client, logger, socialServiceName, socialServiceTag, "GET",
		"/api/social/v1/{id}/reviews", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, socialServiceName, socialServiceTag, "POST",
		"/api/social/v1/{id}/reviews/add", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, socialServiceName, socialServiceTag, "GET",
		"/api/social/v1/{id}/followees", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, socialServiceName, socialServiceTag, "POST",
		"/api/social/v1/{id}/followees/add", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, socialServiceName, socialServiceTag, "DELETE",
		"/api/social/v1/{productid}/{userid}/followees/delete", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, socialServiceName, socialServiceTag, "GET",
		"/api/social/v1/{productid}/summary", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, socialServiceName, socialServiceTag, "GET",
		"/api/social/v1/{user}/followees/products", duration, 3)

	// Add API gateway for user service
	r = NSUtil.RegisterSDService(ctx, r, client, logger, userServiceName, userServiceTag, "GET",
		"/api/v1/register", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, userServiceName, userServiceTag, "POST",
		"/api/v1/authenticate", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, userServiceName, userServiceTag, "GET",
		"/api/v1/users/{username}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, userServiceName, userServiceTag, "POST",
		"/api/v1/users/{username}/update", duration, 3)

	// Add API gateway for proudct Service
	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "POST",
		"/api/upload/style", 4*duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "POST",
		"/api/upload/styles", 10*duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "GET",
		"/api/artists", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "GET",
		"/api/artists/hotest", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "GET",
		"/api/products", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "GET",
		"/api/products/user/{usrid}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "GET",
		"/api/products/tags/{tags}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "GET",
		"/api/products/{id}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "GET",
		"/api/search", 10*duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "GET",
		"/api/v1/cache/get/{usrid}/{imgid}", 3*duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "DELETE",
		"/api/products/{id}/delete", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "POST",
		"/api/products/{id}/update", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, productsServiceName, productsServiceTag, "POST",
		"/api/products/{id}/transactionupdate", duration, 3)

	return r

}
