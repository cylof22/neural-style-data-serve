package OrderService

import (
	"context"
	"github.com/go-kit/kit/endpoint"
)


type NSGetOrdersRequest struct {
	Buyer string
}

type NSOrdersResponse struct {
	Orders   []Order
	Err      error
}

type NSGetSellingsRequest struct {
	Seller string
}

type NSGetOrderByProductIdRequest struct {
	ProductId string
}

type NSGetOrderByProductIdResponse struct {
	Target    Order
	Err       error
}

type NSSellRequest struct {
	SellInfo  Order
}

type NSErrorResponse struct {
	Err       error
}

type NSOrderIdRequest struct {
	OrderId   string
}

type NSBuyRequest struct {
	OrderId   string
	BuyData   BuyInfo
}

type NSChainRequest struct {
	ChainId   string
	Result    string
}

type NSExpressRequest struct {
	OrderId     string
	ExpressData Express
}

type NSAskForReturnRequest struct {
	OrderId     string
	ReturnData  ReturnInfo
}

func MakeNSGetOrdersEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetOrdersRequest)
		orders, err := svc.GetOrders(req.Buyer)
		return NSOrdersResponse{Orders: orders, Err: err}, err
	}
}

func MakeNSGetSellingsEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetSellingsRequest)
		orders, err := svc.GetSellings(req.Seller)
		return NSOrdersResponse{Orders: orders, Err: err}, err
	}
}

func MakeNSGetOrderByProductIdEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetOrderByProductIdRequest)
		order, err := svc.GetOrderByProductId(req.ProductId)
		return NSGetOrderByProductIdResponse{Target: order, Err: err}, err
	}
}

func MakeNSSellEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSSellRequest)
		err := svc.Sell(req.SellInfo)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSStopSellingEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSOrderIdRequest)
		err := svc.StopSelling(req.OrderId)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSBuyEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSBuyRequest)
		err := svc.Buy(req.OrderId, req.BuyData)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSApplyConfirmFromChainEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSChainRequest)
		err := svc.ApplyConfirmFromChain(req.ChainId, req.Result)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSShipProductEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSExpressRequest)
		err := svc.ShipProduct(req.OrderId, req.ExpressData)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSConfirmOrderEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSOrderIdRequest)
		err := svc.ConfirmOrder(req.OrderId)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSAskForReturnEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAskForReturnRequest)
		err := svc.AskForReturn(req.OrderId, req.ReturnData)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSAgreeReturnEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSOrderIdRequest)
		err := svc.AgreeReturn(req.OrderId)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSShipReturnEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSExpressRequest)
		err := svc.ShipReturn(req.OrderId, req.ExpressData)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSConfirmReturnEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSOrderIdRequest)
		err := svc.ConfirmReturn(req.OrderId)
		return NSErrorResponse{Err: err}, err
	}
}

func MakeNSApplyCancelFromChainEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSChainRequest)
		err := svc.ApplyCancelFromChain(req.ChainId, req.Result)
		return NSErrorResponse{Err: err}, err
	}
}