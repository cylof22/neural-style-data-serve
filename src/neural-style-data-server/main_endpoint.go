package main

import (
	"context"
	"encoding/json"
	"net/http"
	"neural-style-util"

	"neural-style-products"

	"neural-style-user"

	"neural-style-transfer"

	"neural-style-order"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"
)

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("context-type", "application/json,charset=utf8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func makeHTTPHandler(ctx context.Context, dbSession *mgo.Session, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(NSUtil.ParseToken),
	}

	authMiddleware := NSUtil.AuthMiddleware(logger)
	// Style Service
	styleTransferService := StyleService.NewNeuralTransferSVC(*networkPath, *previewNetworkPath,
		*outputPath, *serverURL, *serverPort)
	r = StyleService.MakeHTTPHandler(ctx, r, authMiddleware, styleTransferService, options...)

	// Product service
	storageServiceURL := "http://" + *storageServerURL + ":" + *storageServerPort
	storageSaveURL := storageServiceURL + *storageServerSaveRouter
	storageFindURL := storageServiceURL + *storageServerFindRouter

	cacheServiceURL := "http://" + *serverURL + ":" + *serverPort
	cacheGetURL := cacheServiceURL + *cacheGetRouter

	var prods ProductService.Service
	prods = ProductService.NewProductSVC(*outputPath, *serverURL, *serverPort,
		storageSaveURL, storageFindURL, cacheGetURL, *localDev, logger, dbSession)

	prods = ProductService.NewLoggingService(log.With(logger, "component", "product"), prods)
	r = ProductService.MakeHTTPHandler(ctx, r, authMiddleware, prods, options...)

	// User service
	var users UserService.Service
	users = UserService.NewUserSVC(*serverURL, *serverPort, logger, dbSession)
	users = UserService.NewLoggingService(log.With(logger, "component", "product"), users)
	r = UserService.MakeHTTPHandler(ctx, r, authMiddleware, users, options...)

	// Order service
	productsURL := "http://" + *serverURL + ":" + *serverPort + *productsRouter
	var orders OrderService.Service
	orders = OrderService.NewOrderSVC(*serverURL, *serverPort, logger, dbSession, productsURL)
	orders = OrderService.NewLoggingService(log.With(logger, "component", "order"), orders)
	r = OrderService.MakeHTTPHandler(ctx, r, authMiddleware, orders, options...)

	return r
}
