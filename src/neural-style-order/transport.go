package OrderService

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"neural-style-util"

	"github.com/go-kit/kit/endpoint"

	httptransport "github.com/go-kit/kit/transport/http"

	"github.com/gorilla/mux"
)

type ProductQuery struct {
	ProductId      []string `json:"productId"`
}

type ChainResult struct {
	Result      string `json:"result"`
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

func decodeNSGetOrderByProductIdRequest(_ context.Context, r *http.Request) (interface{}, error) {
	queryData, _ := url.ParseQuery(r.URL.RawQuery)
	queryBytes, _ := json.Marshal(queryData)

	var param ProductQuery
	json.Unmarshal(queryBytes, &param)

	return NSGetOrderByProductIdRequest{ProductId: param.ProductId[0]}, nil
}

func encodeNSGetOrderByProductIdResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	result := response.(NSGetOrderByProductIdResponse)
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

func decodeNSOrderIdRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	orderId := vars["id"]

	return NSOrderIdRequest{OrderId: orderId}, nil
}

func decodeNSBuyRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	orderId := vars["id"]

	buyInfo := BuyInfo{}
	json.NewDecoder(r.Body).Decode(&buyInfo)
	return NSBuyRequest{OrderId: orderId, BuyData: buyInfo}, nil
}

func decodeNSChainRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	chainId := vars["chainId"]

	result := ChainResult{}
	json.NewDecoder(r.Body).Decode(&result)
	return NSChainRequest{ChainId: chainId, Result: result.Result}, nil
}

func decodeNSExpressRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	orderId := vars["id"]

	expressData := Express{}
	json.NewDecoder(r.Body).Decode(&expressData)
	return NSExpressRequest{OrderId: orderId, ExpressData: expressData}, nil
}

func decodeNSAskForReturnRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	orderId := vars["id"]

	returnInfo := ReturnInfo{}
	json.NewDecoder(r.Body).Decode(&returnInfo)
	return NSAskForReturnRequest{OrderId: orderId, ReturnData: returnInfo}, nil
}

// MakeHTTPHandler generate the http handler for the style service handler
func MakeHTTPHandler(ctx context.Context, r *mux.Router, auth endpoint.Middleware, svc Service, options ...httptransport.ServerOption) *mux.Router {
	// GET /api/v1/transactionorders
	r.Methods("GET").Path("/api/v1/transactionorders").Handler(httptransport.NewServer(
		MakeNSGetOrdersInTransactionEndpoint(svc),
		decodeNSGetOrdersInTransactionRequest,
		encodeNSOrdersResponse,
		options...,
	))

	// GET /api/v1/orders/{username}
	r.Methods("GET").Path("/api/v1/orders/{username}").Handler(httptransport.NewServer(
		auth(MakeNSGetOrdersEndpoint(svc)),
		decodeNSGetOrdersRequest,
		encodeNSOrdersResponse,
		options...,
	))

	// GET /api/v1/sellings/{username}
	r.Methods("GET").Path("/api/v1/sellings/{username}").Handler(httptransport.NewServer(
		auth(MakeNSGetSellingsEndpoint(svc)),
		decodeNSGetSellingsRequest,
		encodeNSOrdersResponse,
		options...,
	))

	// GET /api/v1/order
	r.Methods("GET").Path("/api/v1/order").Handler(httptransport.NewServer(
		MakeNSGetOrderByProductIdEndpoint(svc),
		decodeNSGetOrderByProductIdRequest,
		encodeNSGetOrderByProductIdResponse,
		options...,
	))
	
	// POST /api/v1/orders/create
	orderCreateHandler := httptransport.NewServer(
		auth(MakeNSSellEndpoint(svc)),
		decodeNSSellRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/order/create").Handler(NSUtil.AccessControl(orderCreateHandler))
	
	// GET /api/v1/orders/{id}/delete
	r.Methods("GET").Path("/api/v1/orders/{id}/delete").Handler(httptransport.NewServer(
		auth(MakeNSStopSellingEndpoint(svc)),
		decodeNSOrderIdRequest,
		encodeNSErrorResponse,
		options...,
	))
	
	// POST /api/v1/orders/{id}/buy
	buyHandler := httptransport.NewServer(
		auth(MakeNSBuyEndpoint(svc)),
		decodeNSBuyRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{id}/buy").Handler(NSUtil.AccessControl(buyHandler))
	
	// POST /api/v1/orders/{chainId}/chainconfirm
	chainApplyHandler := httptransport.NewServer(
		auth(MakeNSApplyConfirmFromChainEndpoint(svc)),
		decodeNSChainRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{chainId}/chainconfirm").Handler(NSUtil.AccessControl(chainApplyHandler))
	
	// POST /api/v1/orders/{id}/productship
	shipProHandler := httptransport.NewServer(
		auth(MakeNSShipProductEndpoint(svc)),
		decodeNSExpressRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{id}/productship").Handler(NSUtil.AccessControl(shipProHandler))
	
	// GET /api/v1/orders/{id}/confirm
	r.Methods("GET").Path("/api/v1/orders/{id}/confirm").Handler(httptransport.NewServer(
		auth(MakeNSConfirmOrderEndpoint(svc)),
		decodeNSOrderIdRequest,
		encodeNSErrorResponse,
		options...,
	))
	
	// POST /api/v1/orders/{id}/askreturn
	askReturnHandler := httptransport.NewServer(
		auth(MakeNSAskForReturnEndpoint(svc)),
		decodeNSAskForReturnRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{id}/askreturn").Handler(NSUtil.AccessControl(askReturnHandler))
	
	// GET /api/v1/orders/{id}/returnagreed
	r.Methods("GET").Path("/api/v1/orders/{id}/returnagreed").Handler(httptransport.NewServer(
		auth(MakeNSAgreeReturnEndpoint(svc)),
		decodeNSOrderIdRequest,
		encodeNSErrorResponse,
		options...,
	))
	
	// POST /api/v1/orders/{id}/returnship
	shipReturnHandler := httptransport.NewServer(
		auth(MakeNSShipReturnEndpoint(svc)),
		decodeNSExpressRequest,
		encodeNSErrorResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/orders/{id}/returnship").Handler(NSUtil.AccessControl(shipReturnHandler))
	
	// GET /api/v1/orders/{id}/returnconfirmed
	r.Methods("GET").Path("/api/v1/orders/{id}/returnconfirmed").Handler(httptransport.NewServer(
		auth(MakeNSConfirmReturnEndpoint(svc)),
		decodeNSOrderIdRequest,
		encodeNSErrorResponse,
		options...,
	))
	
	// POST /api/v1/orders/{chainId}/chaincancel
	chainCancelHandler := httptransport.NewServer(
		auth(MakeNSApplyCancelFromChainEndpoint(svc)),
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