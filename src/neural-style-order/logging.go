package OrderService

import (
	"time"
	"github.com/go-kit/kit/log"
)

type orderService struct {
	logger       log.Logger
	dataService  Service
}

// NewLoggingService returns a new instance of a products logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &orderService{logger, s}
}

func (svc *orderService) GetOrdersInTransaction() (orders []Order, err error){
	defer func(begin time.Time) {
		svc.logger.Log("method", "GetOrdersInTransaction", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetOrdersInTransaction()
}

func (svc *orderService) GetOrders(buyer string) (orders []Order, err error){
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

func (svc *orderService) GetOrderByProductId(productId string) (order Order, err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "GetOrderByProductId", "productId", productId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.GetOrderByProductId(productId)
}

func (svc *orderService) Sell(sellInfo Order) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "Sell", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.Sell(sellInfo)
}

func (svc *orderService) StopSelling(orderId string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "StopSelling", "orderId", orderId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.StopSelling(orderId)
}

func (svc *orderService) Buy(orderId string, buyInfo BuyInfo) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "Buy", "orderId", orderId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.Buy(orderId, buyInfo)
}

func (svc *orderService) ApplyConfirmFromChain(chainId string, result string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ApplyConfirmFromChain", "chainId", chainId, "result", result, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ApplyConfirmFromChain(chainId, result)
}

func (svc *orderService) ShipProduct(orderId string, express Express) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ShipProduct", "orderId", orderId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ShipProduct(orderId, express)
}

func (svc *orderService) ConfirmOrder(orderId string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ConfirmOrder", "orderId", orderId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ConfirmOrder(orderId)
}

func (svc *orderService) AskForReturn(orderId string, returnInfo ReturnInfo) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "AskForReturn", "orderId", orderId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.AskForReturn(orderId, returnInfo)
}

func (svc *orderService) AgreeReturn(orderId string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "AgreeReturn", "orderId", orderId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.AgreeReturn(orderId)
}

func (svc *orderService) ShipReturn(orderId string, express Express) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ShipReturn", "orderId", orderId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ShipReturn(orderId, express)
}

func (svc *orderService) ConfirmReturn(orderId string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ConfirmReturn", "orderId", orderId, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ConfirmReturn(orderId)
}

func (svc *orderService) ApplyCancelFromChain(chainId string, result string) (err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "ApplyCancelFromChain", "chainId", chainId, "result", result, "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.dataService.ApplyCancelFromChain(chainId, result)
}