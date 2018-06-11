package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"neural-style-chain"
	"neural-style-products"
	"neural-style-util"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	orderTable       = "ORDER"
	closedOrderTable = "CLOSEDORDER"
	returnTable      = "RETURN"
	expressTable     = "EXPRESS"
)

var generalErrorInfo = "Server is busy. Please try it later."
var maxDuration = 30
var testDev = true

// OrderStatus define the order status
type OrderStatus struct {
	Status string `json:"status"`
}

// Order define the basic information of the order
type Order struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	PriceType  string `json:"priceType"`
	PriceValue string `json:"priceValue"`

	ProductID string `json:"productID"`

	StartTime       string    `json:"startTime"`
	CompleteTime    time.Time `json:"completeTime"`
	ServerStartTime time.Time `json:"serverStartTime"`
	Duration        string    `json:"duration"`

	ChainID string `json:"chainId"`

	Express Express `json:"express"`
	BuyInfo BuyInfo `json:"buyInfo"`
}

// BuyInfo define the basic buyer information
type BuyInfo struct {
	Buyer           string    `json:"buyer"`
	PriceValue      string    `json:"priceValue"`
	StartTime       string    `json:"startTime"`
	ServerStartTime time.Time `json:"serverStartTime"`
}

// Express define the express information
type Express struct {
	Company   string    `json:"company"`
	Number    string    `json:"number"`
	StartTime time.Time `json:"startTime"`
}

// ReturnInfo define the return request
type ReturnInfo struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	OrderID     string    `json:"orderid"`
	Description string    `json:"description"`
	Images      []string  `json:"images"`
	AskTime     time.Time `json:"askTime"`     // start time when asking return
	AgreeTime   time.Time `json:"agreeTime"`   // start time when agreeing return
	ConfirmTime time.Time `json:"confirmTime"` // start time when confirming return
}

// Service define the basic interface
type Service interface {
	GetOrdersInTransaction() ([]Order, error)
	GetOrders(buyer string) ([]Order, error)
	GetSellings(seller string) ([]Order, error)
	GetOrderByProductID(productID string) (Order, error)
	Sell(sellInfo Order) error
	StopSelling(orderID string) error
	Buy(orderID string, buyInfo BuyInfo) error
	ApplyConfirmFromChain(chainID string, result string) error
	ShipProduct(orderID string, express Express) error
	ConfirmOrder(orderID string) error
	AskForReturn(orderID string, returnInfo ReturnInfo) error
	AgreeReturn(orderID string) error
	ShipReturn(orderID string, express Express) error
	ConfirmReturn(orderID string) error
	ApplyCancelFromChain(chainID string, result string) error
}

// OrderService for order service
type OrderService struct {
	Host        string
	Port        string
	PostDb      *sql.DB
	Logger      log.Logger
	ProductsURL string
}

// NewOrderSVC create a new order service
func NewOrderSVC(host, port string, logger log.Logger, db *sql.DB, productsURL string) Service {
	return &OrderService{Host: host, Port: port, Logger: logger, PostDb: db, ProductsURL: productsURL}
}

// GetOrdersInTransaction get all the orders from the db
func (svc *OrderService) GetOrdersInTransaction() ([]Order, error) {
	rows, err := svc.PostDb.Query("SELECT * FROM orders")
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Find", "Info", err)
		return nil, errors.New(generalErrorInfo)
	}

	var orders []Order

	return orders, nil
}

// GetOrderByProductID get one order by product id
func (svc *OrderService) GetOrderByProductID(productID string) (Order, error) {
	// select from the products table
	row := svc.PostDb.QueryRow("SELECT * FROM " + orderTable + " WHERE PRODUCTID = " + productID)

	var order Order
	if row == nil {
		level.Debug(svc.Logger).Log("Product is in order", "false")
		err := errors.New("Bad Product ID")
		level.Error(svc.Logger).Log("Find Error", err)
		return order, err
	}

	// parse the row

	return order, nil
}

// Sell tags the product information as sell
func (svc *OrderService) Sell(sellInfo Order) error {
	level.Debug(svc.Logger).Log("Input", "productId", "value", sellInfo.ProductID)

	// Query the Product from the Mongodb
	var productInOrder ProductService.Product

	if productInOrder.Type == strconv.Itoa(NSUtil.OnlyShow) {
		level.Error(svc.Logger).Log("PriceType", productInOrder.Type, "Info", "can't be sold")
		return errors.New("Product for Shown can't be sold")
	}

	//order, err := svc.GetOrderByProductID(sellInfo.ProductID)

	sellInfo.Status = strconv.Itoa(NSUtil.None)
	sellInfo.ID = NSUtil.UniqueID()
	sellInfo.ServerStartTime = time.Now()
	inputDuration, err := strconv.Atoi(sellInfo.Duration)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Atoi", "Error", err)
		return errors.New("Please set right duration")
	}

	if inputDuration > maxDuration {
		sellInfo.Duration = strconv.Itoa(maxDuration)
	}

	// update the sell information
	// insert individual element
	insertQuery := "INSERT INFO " + orderTable
	_, err = svc.PostDb.Exec(insertQuery)
	//err = c.Insert(sellInfo)
	if err != nil {
		level.Error(svc.Logger).Log("Insert error", err)
		return errors.New(generalErrorInfo)
	}

	// sending message to chain
	// ? change string to int, only string is not enought
	proTypeString, _ := strconv.Atoi(productInOrder.Type)
	ChainService.StartToSell(sellInfo.ChainID, sellInfo.PriceValue, proTypeString)

	return nil
}

// StopSelling tag the selling product as unselling status
func (svc *OrderService) StopSelling(orderID string) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderID)
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Error", err)
		return errors.New(generalErrorInfo)
	}

	return svc.stopSelling(order)
}

func (svc *OrderService) stopSelling(order Order) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", order.ID)
	if order.Status != strconv.Itoa(NSUtil.None) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be stopped now")
		return errors.New("Can't cancel the order because it's in transaction")
	}

	err := svc.deleteOrder(order.ID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "deleteOrder", "Error", err)
		return errors.New(generalErrorInfo)
	}

	ChainService.StopSelling(order.ChainID)
	return nil
}

// Buy launch the buy request
func (svc *OrderService) Buy(orderID string, buyInfo BuyInfo) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderID)
	// get current order
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Error", err)
		return errors.New(generalErrorInfo)
	}

	var productInOrder ProductService.Product

	// if the product can be bought
	productType := productInOrder.Type
	orderType := order.PriceType

	if productType == strconv.Itoa(NSUtil.Fix) {
		// bought by others
		if order.Status != strconv.Itoa(NSUtil.None) {
			level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be bought")
			return errors.New("The product has been sold. Please try the others")
		}
	} else if productType == strconv.Itoa(NSUtil.Auction) {
		if order.Status != strconv.Itoa(NSUtil.None) &&
			order.Status != strconv.Itoa(NSUtil.InAuction) {
			level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be bought")
			return errors.New("The product has been sold. Please try the others")
		}
	} else {
		level.Error(svc.Logger).Log("PriceType", productType, "Info", "isn't supported now")
		return errors.New("The product can't be bought. Please try the others")
	}

	// get current status and send message to chain if necessary
	var updateStatus = strconv.Itoa(NSUtil.None)
	if orderType == strconv.Itoa(NSUtil.Fix) {
		if productType == strconv.Itoa(NSUtil.Digit) {
			updateStatus = strconv.Itoa(NSUtil.InFix)

			// send the transaction to chain
			err = ChainService.ConfirmOrder(order.ChainID)
			if err != nil {
				level.Error(svc.Logger).Log("API", "Chain.ConfirmOrder", "Info", err)
				return errors.New(generalErrorInfo)
			}
		} else if productType == strconv.Itoa(NSUtil.Entity) {
			updateStatus = strconv.Itoa(NSUtil.Unshipped)
		}
	} else if orderType == strconv.Itoa(NSUtil.Auction) {
		updateStatus = strconv.Itoa(NSUtil.InAuction)

		err = ChainService.UpdatePrice(order.ChainID, buyInfo.Buyer, order.PriceValue)
		if err != nil {
			level.Error(svc.Logger).Log("API", "Chain.UpdatePrice", "Info", err)
			return errors.New(generalErrorInfo)
		}
	} else {
		level.Error(svc.Logger).Log("PriceType", "Chain.UpdatePrice", "Info", err)
		return errors.New("The product can't be bought. Please try the others")
	}

	// insert the buy infor by buy table
	buyInfo.ServerStartTime = time.Now()

	// update the order data by orderid in order table
	updateQuery := "UPDATE " + orderTable + " SET " +
		"STATUS=" + updateStatus + " WHERE " + "ORDERID=" + orderID

	_, err = svc.PostDb.Exec(updateQuery)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Date.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if testDev {
		if orderType == strconv.Itoa(NSUtil.Fix) {
			svc.ApplyConfirmFromChain(orderID, "success")
		}
	}

	return nil
}

// ApplyConfirmFromChain confirm the chain transction and update the owner
func (svc *OrderService) ApplyConfirmFromChain(chainID string, result string) error {
	// get order
	order, err := svc.getOrderByChainID(chainID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderByChainId", "Info", err)
		return errors.New(generalErrorInfo)
	}

	var productInOrder ProductService.Product

	productType := productInOrder.Type
	orderType := order.PriceType

	// if the order can be completed
	if orderType == strconv.Itoa(NSUtil.Fix) {
		if productType == strconv.Itoa(NSUtil.Digit) {
			if order.Status != strconv.Itoa(NSUtil.InFix) {
				level.Error(svc.Logger).Log("PriceType", orderType, "Status", order.Status, "Info", "can't be completed")
				return errors.New("Current order can't be completed")
			}
		} else if productType == strconv.Itoa(NSUtil.Entity) {
			if order.Status != strconv.Itoa(NSUtil.DispatchConfirmed) {
				level.Error(svc.Logger).Log("PriceType", orderType, "Status", order.Status, "Info", "can't be completed")
				return errors.New("Current order can't be completed")
			}
		} else {
			level.Error(svc.Logger).Log("PriceType", orderType, "Info", "unsupported")
			return errors.New("Current order isn't in a transaction")
		}
	} else if orderType == strconv.Itoa(NSUtil.Auction) {
		if productType == strconv.Itoa(NSUtil.Digit) {
			if order.Status != strconv.Itoa(NSUtil.InAuction) {
				level.Error(svc.Logger).Log("PriceType", orderType, "Status", order.Status, "Info", "can't be completed")
				return errors.New("Current order can't be completed")
			}
		} else if productType == strconv.Itoa(NSUtil.Entity) {
			if order.Status != strconv.Itoa(NSUtil.DispatchConfirmed) {
				level.Error(svc.Logger).Log("PriceType", orderType, "Status", order.Status, "Info", "can't be completed")
				return errors.New("Current order can't be completed")
			}
		} else {
			level.Error(svc.Logger).Log("PriceType", orderType, "Info", "unsupported")
			return errors.New("Current order isn't in a transaction")
		}
	} else {
		level.Error(svc.Logger).Log("PriceType", orderType, "Info", "unsupported")
		return errors.New("Current order isn't in a transaction")
	}

	// get new status
	var newStatus = NSUtil.None
	if result == "fail" {
		// post message to buyer
		newStatus = NSUtil.Failed
	} else if result == "success" {
		newStatus = NSUtil.Completed
	} else {
		level.Error(svc.Logger).Log("ReturnValue", result, "Info", "unsupported")
		return errors.New("unhandled return value")
	}
	order.Status = strconv.Itoa(newStatus)

	// close the order
	err = svc.closeOrder(order)
	if err != nil {
		level.Error(svc.Logger).Log("API", "closeOrder", "Info", err)
		return errors.New(generalErrorInfo)
	}

	// update product owner
	if result == "success" {
		svc.updateProductAfterTransaction(order.ProductID, order.BuyInfo.Buyer, order.BuyInfo.PriceValue)
	}

	return nil
}

func (svc *OrderService) getOrderByChainID(chainID string) (Order, error) {
	var order Order
	row := svc.PostDb.QueryRow("SELECT * FROM orders WHERE CHAINID=" + chainID)

	if row == nil {
		err := errors.New("No order for ChainID = " + chainID)
		level.Error(svc.Logger).Log("API", "Date.find", "Info", err)
		return order, errors.New("can't get order for the chain id " + chainID)
	}

	// parse the order
	return order, nil
}

// OrderIsDue checking the order status is expired
func (svc *OrderService) OrderIsDue(orderID string) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderID)
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	var productInOrder ProductService.Product
	// Query the product from the MongoDB

	orderType := order.PriceType
	// for fix price
	if orderType == strconv.Itoa(NSUtil.Fix) {
		return svc.stopSelling(order)
	} else if orderType == strconv.Itoa(NSUtil.Auction) {
		return svc.auctionIsDue(order, productInOrder)
	} else {
		level.Error(svc.Logger).Log("PriceType", orderType, "Info", "Unsupported")
		return errors.New("The product isn't in transaction")
	}
}

func (svc *OrderService) auctionIsDue(order Order, productInOrder ProductService.Product) error {
	if productInOrder.Type == strconv.Itoa(NSUtil.Digit) {
		return ChainService.ConfirmOrder(order.ChainID)
	} else if productInOrder.Type == strconv.Itoa(NSUtil.Entity) {
		err := svc.updateOrderStatus(order.ID, NSUtil.Unshipped)
		if err != nil {
			level.Error(svc.Logger).Log("API", "updateOrderStatus", "Info", err)
			return errors.New("Failed to stop auction")
		}
	} else {
		level.Error(svc.Logger).Log("ProductType", productInOrder.Type, "Info", "Unsupported")
		return errors.New("unsupported product type for the order " + order.ID)
	}

	return nil
}

// ShipProduct add the order id and its corresponding express partner
func (svc *OrderService) ShipProduct(orderID string, express Express) error {
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderByID", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.Unshipped) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be shipped")
		return errors.New("Current operation isn't supported. Please check order's status")
	}

	express.StartTime = time.Now()
	// Insert the express data

	// update the order table
	// Todo: add express foregin key
	updateQuery := "UPDATE " + orderTable + "SET " + "STATUS=" + strconv.Itoa(NSUtil.Dispatched) +
		" WHERE " + "ORDERID=" + orderID

	_, err = svc.PostDb.Exec(updateQuery)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

func (svc *OrderService) getExpressURL(express Express) string {
	// TODO: get link for the express
	var tempValue = express.Company + ":" + express.Number
	return tempValue
}

// ConfirmOrder confirm the order information
func (svc *OrderService) ConfirmOrder(orderID string) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderID)
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.Dispatched) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be confirmed")
		return errors.New("Current operation isn't supported. Please check order's status")
	}

	err = svc.updateOrderStatus(orderID, NSUtil.DispatchConfirmed)
	if err != nil {
		level.Error(svc.Logger).Log("API", "updateOrderStatus", "Info", err)
		return errors.New(generalErrorInfo)
	}

	ChainService.ConfirmOrder(order.ChainID)

	if testDev {
		if order.PriceType == strconv.Itoa(NSUtil.Fix) {
			svc.ApplyConfirmFromChain(orderID, "success")
		}
	}

	return nil
}

// AskForReturn launch the return request
func (svc *OrderService) AskForReturn(orderID string, returnInfo ReturnInfo) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderID)
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.Dispatched) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be returned")
		return errors.New("Current operation isn't supported. Please check order's status")
	}

	savedReturnInfo := svc.convertReturnInfo(order.BuyInfo.Buyer, returnInfo)
	savedReturnInfo.AskTime = time.Now()
	savedReturnInfo.Status = strconv.Itoa(NSUtil.ReturnInAgree)
	savedReturnInfo.OrderID = orderID

	// insert the return infor to the return table
	insertQuery := "INSERT INFO " + orderTable
	_, err = svc.PostDb.Exec(insertQuery)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

// save returninfo's image from data to a file
func (svc *OrderService) convertReturnInfo(buyer string, returnInfo ReturnInfo) ReturnInfo {
	var savedReturnInfo = ReturnInfo{Description: returnInfo.Description}
	for index, pic := range returnInfo.Images {
		picID := NSUtil.UniqueID()
		picURL, err := NSUtil.UploadPicture(buyer, pic, picID, "returns", true)
		if err == nil {
			savedReturnInfo.Images[index] = picURL
		} else {
			savedReturnInfo.Images[index] = ""
		}
	}

	return savedReturnInfo
}

// AgreeReturn agree the return request of the buyer
func (svc *OrderService) AgreeReturn(orderID string) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderID)
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderByID", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.ReturnInAgree) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be agreed")
		return errors.New("Current operation isn't supported. Please check order's status")
	}

	updateQuery := "UPDATE " + returnTable + " SET STATUS=" + strconv.Itoa(NSUtil.ReturnAgreed) +
		" AGREETIME=" + time.Now().String() + " WHERE ORDERID=" + orderID

	_, err = svc.PostDb.Exec(updateQuery)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

// ShipReturn launch the shipping process of the return request
func (svc *OrderService) ShipReturn(orderID string, express Express) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderID)
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderByID", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.ReturnAgreed) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be shipped")
		return errors.New("Current operation isn't supported. Please check order's status")
	}

	// Insert the express to the express table and get the id
	express.StartTime = time.Now()

	// update the status by orderid in return table
	updateQuery := "UPDATE " + returnTable + " SET " +
		"STATUS=" + strconv.Itoa(NSUtil.ReturnDispatched) +
		" WHERE " + "ORDERID=" + orderID

	_, err = svc.PostDb.Exec(updateQuery)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

// ConfirmReturn the seller accept the return request
func (svc *OrderService) ConfirmReturn(orderID string) error {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderID)
	order, err := svc.getOrderByID(orderID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.ReturnDispatched) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be confirmed")
		return errors.New("Current operation isn't supported. Please check order's status")
	}

	ChainService.CancelOrder(order.ChainID)

	// update data in database
	updateQuery := "UPDATE " + orderTable + " SET " +
		" CONFIRMTIME=" + time.Now().String() +
		" STATUS=" + strconv.Itoa(NSUtil.ReturnConfirmed) + " WHERE ID=" + orderID

	_, err = svc.PostDb.Exec(updateQuery)

	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if testDev {
		if order.PriceType == strconv.Itoa(NSUtil.Fix) {
			svc.ApplyCancelFromChain(orderID, "success")
		}
	}

	return nil
}

// ApplyCancelFromChain launch the cancel request for the blockchain
func (svc *OrderService) ApplyCancelFromChain(chainID string, result string) error {
	level.Debug(svc.Logger).Log("Input", "chainId", "Value", chainID)
	order, err := svc.getOrderByChainID(chainID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderByChainId", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.ReturnConfirmed) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be cancelled")
		return errors.New("Current operation isn't supported. Please check order's status")
	}

	if result != "success" {
		return errors.New("unhandled return value from chain")
	}

	order.Status = strconv.Itoa(NSUtil.ReturnCompleted)
	err = svc.closeOrder(order)
	if err != nil {
		level.Error(svc.Logger).Log("API", "closeOrder", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

func (svc *OrderService) updateOrderStatus(orderID string, orderStatus int) error {
	updateQuery := "UPDATE orders SET STATUS=" + strconv.Itoa(orderStatus) + "WHERE ID=" + orderID
	_, err := svc.PostDb.Exec(updateQuery)

	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New("Failed to complete order")
	}

	return nil
}

// move the order to closed collection
func (svc *OrderService) closeOrder(order Order) error {
	level.Debug(svc.Logger).Log("func", "closeOrder")

	order.CompleteTime = time.Now()

	insertQuery := "INSERT INTO " + closedOrderTable + "VALUES (" +
		order.ID + ")"

	_, err := svc.PostDb.Exec(insertQuery)

	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Insert", "Info", err)
		return errors.New("Failed to close the order")
	}

	err = svc.deleteOrder(order.ID)
	if err != nil {
		return errors.New("Failed to delete completed order")
	}

	return nil
}

// GetOrders get all the orders for the given buyer
func (svc *OrderService) GetOrders(buyer string) ([]Order, error) {
	level.Debug(svc.Logger).Log("Input", "buyer", "Value", buyer)

	rows, err := svc.PostDb.Query("SELECT * FROM orders WHERE BUYER=" + buyer)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Find", "Info", err)
		return nil, errors.New(generalErrorInfo)
	}

	var orders []Order
	// parse the orders from the rows
	return orders, nil
}

// GetSellings get all the selling products from the given seller
func (svc *OrderService) GetSellings(seller string) ([]Order, error) {
	level.Debug(svc.Logger).Log("Input", "seller", "Value", seller)

	rows, err := svc.PostDb.Query("SELECT * FROM orders WHERE SELLER=" + seller)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Find", "Info", err)
		return nil, errors.New(generalErrorInfo)
	}

	var orders []Order
	// parse the order

	return orders, nil
}

func (svc *OrderService) getOrderByID(orderID string) (Order, error) {
	level.Debug(svc.Logger).Log("Func", "getOrderByID")

	var order Order
	row := svc.PostDb.QueryRow("SELECT * FROM orders WHERE ORDERID=" + orderID)
	if row == nil {
		err := errors.New("Bad Order ID")
		level.Error(svc.Logger).Log("API", "Data.Find", "Info", err)
		return order, errors.New("Failed to get order")
	}

	// parse the row

	return order, nil
}

func (svc *OrderService) deleteOrder(orderID string) error {
	level.Debug(svc.Logger).Log("Func", "deleteOrder", "OrderID", orderID)

	deleteQuery := "DELETE FROM " + orderTable + " WHERE ID=" + orderID

	_, err := svc.PostDb.Exec(deleteQuery)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Remove", "Info", err)
		return errors.New("Failed to delete order")
	}

	return nil
}

func (svc *OrderService) updateProductAfterTransaction(productID string, newOwner string, price string) error {
	updateData := NSUtil.TransactionUpdateData{}
	updateData.Owner = newOwner
	updateData.Price = price
	updateString, _ := json.Marshal(updateData)

	updateClient := &http.Client{}
	updateURL := svc.ProductsURL + "/" + productID + "/transactionupdate"
	updateReq, err := http.NewRequest("POST", updateURL, bytes.NewReader(updateString))
	if err != nil {
		level.Error(svc.Logger).Log("Storage", updateURL, "err", err)
		return err
	}

	res, err := updateClient.Do(updateReq)
	if err != nil {
		level.Error(svc.Logger).Log("API", "http.Client.Do", "Error", err)
		return err
	}

	if res.StatusCode != http.StatusOK {
		return errors.New("Product owner update fails")
	}

	return nil
}

func (svc *OrderService) parseRow(row sql.Row) Order {
	var orderInfo Order
	return orderInfo
}
