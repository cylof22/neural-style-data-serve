package main

import (
	"context"
	"encoding/json"
	"net/http"

	"neural-style-products"

	"neural-style-user"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	mgo "gopkg.in/mgo.v2"

	"neural-style-transfer"
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
	}

	// Style Service
	styleTransferService := StyleService.NeuralTransferService{
		NetworkPath:        *networkPath,
		PreviewNetworkPath: *previewNetworkPath,
		OutputPath:         *outputPath,
		Host:               *serverURL,
		Port:               *serverPort,
		Session:            dbSession,
	}
	r = StyleService.MakeHTTPHandler(ctx, r, styleTransferService, options...)

	// Product service
	productService := ProductService.ProductService{
		OutputPath: *outputPath,
		Host:       *serverURL,
		Port:       *serverPort,
		Session:    dbSession,
	}

	r = ProductService.MakeHTTPHandler(ctx, r, productService, options...)

	// User service
	userService := UserService.UserService{
		Host:    *serverURL,
		Port:    *serverPort,
		Session: dbSession,
	}
	r = UserService.MakeHTTPHandler(ctx, r, userService, options...)

	return r
}
