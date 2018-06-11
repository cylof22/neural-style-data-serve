package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
)

// NSGetOrdersRequest define params for getting orders for a given buyer
type NSGetOrdersRequest struct {
	Buyer string
}

// NSOrdersResponse returns all the orders for a given buyer
type NSOrdersResponse struct {
	Orders []Order
	Err    error
}

// NSGetSellingsRequest define the basic information fo the selling
type NSGetSellingsRequest struct {
	Seller string
}

// NSGetOrderByProductIDRequest define the params for get order by product id
type NSGetOrderByProductIDRequest struct {
	ProductID string
}

// NSGetOrderByProductIDResponse returns the order for a given product id
type NSGetOrderByProductIDResponse struct {
	Target Order
	Err    error
}

// NSSellRequest define the basic information for launching a selling request
type NSSellRequest struct {
	SellInfo Order
}

// NSErrorResponse define the basic response for a given error
type NSErrorResponse struct {
	Err error
}

// NSOrderIDRequest define the basic id request
type NSOrderIDRequest struct {
	OrderID string
}

// NSBuyRequest define the basic parameter for a buy request
type NSBuyRequest struct {
	OrderID string
	BuyData BuyInfo
}

// NSChainRequest define the chainid and result information for a chain request
type NSChainRequest struct {
	ChainID string
	Result  string
}

// NSExpressRequest define the basic params for a express request
type NSExpressRequest struct {
	OrderID     string
	ExpressData Express
}

// NSAskForReturnRequest define the basic return request params
type NSAskForReturnRequest struct {
	OrderID    string
	ReturnData ReturnInfo
}

func makeNSGetOrdersEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetOrdersRequest)
		orders, err := svc.GetOrders(req.Buyer)
		return NSOrdersResponse{Orders: orders, Err: err}, err
	}
}

func makeNSGetOrdersInTransactionEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		orders, err := svc.GetOrdersInTransaction()
		return NSOrdersResponse{Orders: orders, Err: err}, err
	}
}

func makeNSGetSellingsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetSellingsRequest)
		orders, err := svc.GetSellings(req.Seller)
		return NSOrdersResponse{Orders: orders, Err: err}, err
	}
}

func makeNSGetOrderByProductIDEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetOrderByProductIDRequest)
		order, err := svc.GetOrderByProductID(req.ProductID)
		return NSGetOrderByProductIDResponse{Target: order, Err: err}, err
	}
}

func makeNSSellEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSSellRequest)
		err := svc.Sell(req.SellInfo)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSStopSellingEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSOrderIDRequest)
		err := svc.StopSelling(req.OrderID)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSBuyEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSBuyRequest)
		err := svc.Buy(req.OrderID, req.BuyData)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSApplyConfirmFromChainEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSChainRequest)
		err := svc.ApplyConfirmFromChain(req.ChainID, req.Result)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSShipProductEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSExpressRequest)
		err := svc.ShipProduct(req.OrderID, req.ExpressData)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSConfirmOrderEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSOrderIDRequest)
		err := svc.ConfirmOrder(req.OrderID)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSAskForReturnEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAskForReturnRequest)
		err := svc.AskForReturn(req.OrderID, req.ReturnData)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSAgreeReturnEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSOrderIDRequest)
		err := svc.AgreeReturn(req.OrderID)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSShipReturnEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSExpressRequest)
		err := svc.ShipReturn(req.OrderID, req.ExpressData)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSConfirmReturnEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSOrderIDRequest)
		err := svc.ConfirmReturn(req.OrderID)
		return NSErrorResponse{Err: err}, err
	}
}

func makeNSApplyCancelFromChainEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSChainRequest)
		err := svc.ApplyCancelFromChain(req.ChainID, req.Result)
		return NSErrorResponse{Err: err}, err
	}
}

func encodeError(ctx context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("context-type", "application/json,charset=utf8")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
