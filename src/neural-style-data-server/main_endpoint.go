package main

import (
	"context"
	"net/http"
	"neural-style-util"
	"time"

	"neural-style-user"

	"neural-style-transfer"

	"neural-style-order"

	"github.com/go-kit/kit/log"
	consulsd "github.com/go-kit/kit/sd/consul"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
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

	// User service
	var users UserService.Service
	users = UserService.NewUserSVC(*serverURL, *serverPort, logger, dbSession)
	users = UserService.NewLoggingService(log.With(logger, "component", "user"), users)
	r = UserService.MakeHTTPHandler(ctx, r, authMiddleware, users, options...)

	// Order service
	productsURL := "http://" + *serverURL + ":" + *serverPort + *productsRouter
	var orders OrderService.Service
	orders = OrderService.NewOrderSVC(*serverURL, *serverPort, logger, dbSession, productsURL)
	orders = OrderService.NewLoggingService(log.With(logger, "component", "order"), orders)
	r = OrderService.MakeHTTPHandler(ctx, r, authMiddleware, orders, options...)

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
	// Add API gateway for proudct Service
	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "upload-style", "POST",
		"/api/upload/style", 4*duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "upload-styles", "POST",
		"/api/upload/styles", 10*duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "get-artists", "GET",
		"/api/artists", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "get-hotest-artists", "GET",
		"/api/artists/hotest", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "get-products", "GET",
		"/api/products", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "get-user-products", "GET",
		"/api/products/user/{usrid}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "get-tags", "GET",
		"/api/products/tags/{tags}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "prouducts", "get-user-product", "GET",
		"/api/products/{id}", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "get-search", "GET",
		"/api/search", 10*duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "get-cached-image", "GET",
		"/api/v1/cache/get/{usrid}/{imgid}", 3*duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "delete-product", "DELETE",
		"/api/products/{id}/delete", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "update-product", "POST",
		"/api/products/{id}/update", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "products", "post-transaction", "POST",
		"/api/products/{id}/transactionupdate", duration, 3)

	// Add API gateway for Social Service
	r = NSUtil.RegisterSDService(ctx, r, client, logger, "social", "get-reviews", "GET",
		"/api/social/v1/{id}/reviews", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "social", "add-review", "POST",
		"/api/social/v1/{id}/reviews/add", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "social", "get-followees", "GET",
		"/api/social/v1/{id}/followees", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "social", "add-followee", "POST",
		"/api/social/v1/{id}/followees/add", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "social", "delete-followee", "DELETE",
		"/api/social/v1/{productid}/{userid}/followees/delete", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "social", "get-summary", "GET",
		"/api/social/v1/{productid}/summary", duration, 3)

	r = NSUtil.RegisterSDService(ctx, r, client, logger, "social", "get-summary", "GET",
		"/api/social/v1/{user}/followees/products", duration, 3)

	return r

}
