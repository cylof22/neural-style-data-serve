package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/url"

	"neural-style-util"

	"github.com/go-kit/kit/log"
	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/gorilla/mux"
)

// ProductQuery define the query ID
type ProductQuery struct {
	ProductID []string `json:"productId"`
}

// ChainResult define the update result from the blockchain
type ChainResult struct {
	Result string `json:"result"`
}

func decodeNSGetOrdersRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	username := vars["username"]

	return NSGetOrdersRequest{Buyer: username}, nil
}

func encodeNSOrdersResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	ordersRes := response.(NSOrdersResponse)
	if ordersRes.Err != nil {
		return ordersRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(ordersRes.Orders)
}

func decodeNSGetOrdersInTransactionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeNSGetSellingsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	username := vars["username"]

	return NSGetSellingsRequest{Seller: username}, nil
}

func decodeNSGetOrderByProductIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	queryData, _ := url.ParseQuery(r.URL.RawQuery)
	queryBytes, _ := json.Marshal(queryData)

	var param ProductQuery
	json.Unmarshal(queryBytes, &param)

	return NSGetOrderByProductIDRequest{ProductID: param.ProductID[0]}, nil
}

func encodeNSGetOrderByProductIDResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	result := response.(NSGetOrderByProductIDResponse)
	if result.Err != nil {
		return result.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(result.Target)
}

func decodeNSSellRequest(_ context.Context, r *http.Request) (interface{}, error) {
	order := Order{}
	json.NewDecoder(r.Body).Decode(&order)
	return NSSellRequest{SellInfo: order}, nil
}

func encodeNSErrorResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	result := response.(NSErrorResponse)
	if result.Err != nil {
		return result.Err
	}

	return nil
}

func decodeNSOrderIDRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	return NSOrderIDRequest{OrderID: orderID}, nil
}

func decodeNSBuyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	buyInfo := BuyInfo{}
	json.NewDecoder(r.Body).Decode(&buyInfo)
	return NSBuyRequest{OrderID: orderID, BuyData: buyInfo}, nil
}

func decodeNSChainRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	chainID := vars["chainId"]

	result := ChainResult{}
	json.NewDecoder(r.Body).Decode(&result)
	return NSChainRequest{ChainID: chainID, Result: result.Result}, nil
}

func decodeNSExpressRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	expressData := Express{}
	json.NewDecoder(r.Body).Decode(&expressData)
	return NSExpressRequest{OrderID: orderID, ExpressData: expressData}, nil
}

func decodeNSAskForReturnRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	returnInfo := ReturnInfo{}
	json.NewDecoder(r.Body).Decode(&returnInfo)
	return NSAskForReturnRequest{OrderID: orderID, ReturnData: returnInfo}, nil
}

func makeHTTPHandler(ctx context.Context, db *sql.DB, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	options := []httptransport.ServerOption{
		httptransport.ServerErrorLogger(logger),
		httptransport.ServerErrorEncoder(encodeError),
		httptransport.ServerBefore(NSUtil.ParseToken),
	}

	auth := NSUtil.AuthMiddleware(logger)

	// Order service
	productsURL := "http://" + *serverURL + ":" + *serverPort + *productsRouter
	svc := NewOrderSVC(*serverURL, *serverPort, logger, db, productsURL)
	svc = NewLoggingService(log.With(logger, "component", "order"), svc)

	// GET /api/v1/transactionorders
	r.Methods("GET").Path("/api/v1/transactionorders").Handler(httptransport.NewServer(
		makeNSGetOrdersInTransactionEndpoint(svc),
		decodeNSGetOrdersInTransactionRequest,
		encodeNSOrdersResponse,
		options...,
	))

	// GET /api/v1/orders/{username}
	r.Methods("GET").Path("/api/v1/orders/{username}").Handler(httptransport.NewServer(
		auth(makeNSGetOrdersEndpoint(svc)),
		decodeNSGetOrdersRequest,
		encodeNSOrdersResponse,
		options...,
	))

	// GET /api/v1/sellings/{username}
	r.Methods("GET").Path("/api/v1/sellings/{username}").Handler(httptransport.NewServer(
		auth(makeNSGetSellingsEndpoint(svc)),
		decodeNSGetSellingsRequest,
		encodeNSOrdersResponse,
		options...,
	))

	// GET /api/v1/order
	r.Methods("GET").Path("/api/v1/order").Handler(httptransport.NewServer(
		makeNSGetOrderByProductIDEndpoint(svc),
		decodeNSGetOrderByProductIDRequest,
		encodeNSGetOrderByProductIDResponse,
		options...,
	))

	// POST /api/v1/orders/create
	orderCreateHandler := httptransport.NewServer(
		auth(makeNSSellEndpoint(svc)),
		decodeNSSellRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/order/create").Handler(NSUtil.AccessControl(orderCreateHandler))

	// GET /api/v1/orders/{id}/delete
	r.Methods("GET").Path("/api/v1/orders/{id}/delete").Handler(httptransport.NewServer(
		auth(makeNSStopSellingEndpoint(svc)),
		decodeNSOrderIDRequest,
		encodeNSErrorResponse,
		options...,
	))

	// POST /api/v1/orders/{id}/buy
	buyHandler := httptransport.NewServer(
		auth(makeNSBuyEndpoint(svc)),
		decodeNSBuyRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{id}/buy").Handler(NSUtil.AccessControl(buyHandler))

	// POST /api/v1/orders/{chainId}/chainconfirm
	chainApplyHandler := httptransport.NewServer(
		auth(makeNSApplyConfirmFromChainEndpoint(svc)),
		decodeNSChainRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{chainId}/chainconfirm").Handler(NSUtil.AccessControl(chainApplyHandler))

	// POST /api/v1/orders/{id}/productship
	shipProHandler := httptransport.NewServer(
		auth(makeNSShipProductEndpoint(svc)),
		decodeNSExpressRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{id}/productship").Handler(NSUtil.AccessControl(shipProHandler))

	// GET /api/v1/orders/{id}/confirm
	r.Methods("GET").Path("/api/v1/orders/{id}/confirm").Handler(httptransport.NewServer(
		auth(makeNSConfirmOrderEndpoint(svc)),
		decodeNSOrderIDRequest,
		encodeNSErrorResponse,
		options...,
	))

	// POST /api/v1/orders/{id}/askreturn
	askReturnHandler := httptransport.NewServer(
		auth(makeNSAskForReturnEndpoint(svc)),
		decodeNSAskForReturnRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{id}/askreturn").Handler(NSUtil.AccessControl(askReturnHandler))

	// GET /api/v1/orders/{id}/returnagreed
	r.Methods("GET").Path("/api/v1/orders/{id}/returnagreed").Handler(httptransport.NewServer(
		auth(makeNSAgreeReturnEndpoint(svc)),
		decodeNSOrderIDRequest,
		encodeNSErrorResponse,
		options...,
	))

	// POST /api/v1/orders/{id}/returnship
	shipReturnHandler := httptransport.NewServer(
		auth(makeNSShipReturnEndpoint(svc)),
		decodeNSExpressRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{id}/returnship").Handler(NSUtil.AccessControl(shipReturnHandler))

	// GET /api/v1/orders/{id}/returnconfirmed
	r.Methods("GET").Path("/api/v1/orders/{id}/returnconfirmed").Handler(httptransport.NewServer(
		auth(makeNSConfirmReturnEndpoint(svc)),
		decodeNSOrderIDRequest,
		encodeNSErrorResponse,
		options...,
	))

	// POST /api/v1/orders/{chainId}/chaincancel
	chainCancelHandler := httptransport.NewServer(
		auth(makeNSApplyCancelFromChainEndpoint(svc)),
		decodeNSChainRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{chainId}/chaincancel").Handler(NSUtil.AccessControl(chainCancelHandler))

	// images for explaining return
	returnFiles := http.FileServer(http.Dir("data/returns"))
	r.PathPrefix("/returns/").Handler(http.StripPrefix("/returns/", returnFiles))

	return r
}
