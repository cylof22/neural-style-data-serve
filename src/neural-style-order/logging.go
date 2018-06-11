package main

import (
	"time"

	"github.com/go-kit/kit/log"
)

type orderService struct {
	logger      log.Logger
	dataService Service
}

// NewLoggingService returns a new instance of a products logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &orderService{logger, s}
}

func (svc *orderService) GetOrdersInTransaction() (orders []Order, err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "GetOrdersInTransaction", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetOrdersInTransaction()
}

func (svc *orderService) GetOrders(buyer string) (orders []Order, err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "GetOrders", "buyer", buyer, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetOrders(buyer)
}

func (svc *orderService) GetSellings(seller string) (orders []Order, err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "GetSellings", "seller", seller, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetSellings(seller)
}

func (svc *orderService) GetOrderByProductID(productID string) (order Order, err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "GetOrderByProductId", "productId", productID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetOrderByProductID(productID)
}

func (svc *orderService) Sell(sellInfo Order) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "Sell", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.Sell(sellInfo)
}

func (svc *orderService) StopSelling(orderID string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "StopSelling", "orderId", orderID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.StopSelling(orderID)
}

func (svc *orderService) Buy(orderID string, buyInfo BuyInfo) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "Buy", "orderId", orderID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.Buy(orderID, buyInfo)
}

func (svc *orderService) ApplyConfirmFromChain(chainID string, result string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ApplyConfirmFromChain", "chainId", chainID, "result", result, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ApplyConfirmFromChain(chainID, result)
}

func (svc *orderService) ShipProduct(orderID string, express Express) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ShipProduct", "orderId", orderID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ShipProduct(orderID, express)
}

func (svc *orderService) ConfirmOrder(orderID string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ConfirmOrder", "orderId", orderID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ConfirmOrder(orderID)
}

func (svc *orderService) AskForReturn(orderID string, returnInfo ReturnInfo) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "AskForReturn", "orderId", orderID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.AskForReturn(orderID, returnInfo)
}

func (svc *orderService) AgreeReturn(orderID string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "AgreeReturn", "orderId", orderID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.AgreeReturn(orderID)
}

func (svc *orderService) ShipReturn(orderID string, express Express) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ShipReturn", "orderId", orderID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ShipReturn(orderID, express)
}

func (svc *orderService) ConfirmReturn(orderID string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ConfirmReturn", "orderId", orderID, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ConfirmReturn(orderID)
}

func (svc *orderService) ApplyCancelFromChain(chainID string, result string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ApplyCancelFromChain", "chainId", chainID, "result", result, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ApplyCancelFromChain(chainID, result)
}
