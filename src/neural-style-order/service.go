package OrderService

import (
	"errors"
	"strconv"
	"time"
	"bytes"
	"encoding/json"
	"net/http"
	"neural-style-util"
	"neural-style-chain"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	mgo "gopkg.in/mgo.v2"

	"gopkg.in/mgo.v2/bson"
)

var generalErrorInfo = "Server is busy. Please try it later."
var maxDuration = 30
var testDev bool = false;
type OrderStatus struct {
	Status            string   `json:"status"`
}

type Order struct {
	ID               string          `json:"id"`
	Product          ProductInfo     `json:"product"`
	ChainId          string          `json:"chainId"`
	Status           string          `json:"status"`  
	StartTime        string          `json:"startTime"` 
	ServerStartTime  time.Time       `json:"serverStartTime"`
	Duration         string          `json:"duration"`
	Express          Express         `json:"express"`
	ReturnInfo       ReturnInfo      `json:"returnInfo"`
	BuyInfo          BuyInfo         `json:"buyInfo"`
	CompleteTime     time.Time       `json:"completeTime"`
}

type ProductInfo struct {
	Id               string     `json:"id"`
	Owner            string     `json:"owner"`
	Url              string     `json:"url"`
	Type             string     `json:"type"`
	PriceType        string     `json:"priceType"`
	PriceValue       string     `json:"priceValue"`
}

type BuyInfo struct {
	Buyer            string     `json:"buyer"`
	PriceValue       string     `json:"priceValue"`
	StartTime        string     `json:"startTime"`
	ServerStartTime  time.Time  `json:"serverStartTime"`
}

type Express struct {
	Company          string     `json:"company"`
	Number           string     `json:"number"`
	StartTime        time.Time  `json:"startTime"`
}

type ReturnInfo struct {
	Description      string     `json:"description"`
	Images           []string   `json:"images"`
	AskTime          time.Time  `json:"askTime"`				// start time when asking return
	AgreeTime        time.Time  `json:"agreeTime"`              // start time when agreeing return
	ConfirmTime      time.Time  `json:"confirmTime"`            // start time when confirming return
	Express          Express    `json:"express"`
}

// Service define the basic interface
type Service interface {
	GetOrdersInTransaction() ([]Order, error)
	GetOrders(buyer string) ([]Order, error)
	GetSellings(seller string) ([]Order, error)
	GetOrderByProductId(productId string) (Order, error)
	Sell(sellInfo Order) (error)
	StopSelling(orderId string) (error)
	Buy(orderId string, buyInfo BuyInfo) (error)
	ApplyConfirmFromChain(chainId string, result string) (error)
	ShipProduct(orderId string, express Express) (error)
	ConfirmOrder(orderId string) (error)
	AskForReturn(orderId string, returnInfo ReturnInfo) (error)
	AgreeReturn(orderId string) (error)
	ShipReturn(orderId string, express Express) (error)
	ConfirmReturn(orderId string) (error)
	ApplyCancelFromChain(chainId string, result string) (error)
}

// OrderService for order service
type OrderService struct {
	Host        string
	Port        string
	Session     *mgo.Session
	Logger      log.Logger
	ProductsURL string
}

// NewUserSVC create a new user service
func NewOrderSVC(host, port string, logger log.Logger, session *mgo.Session, productsURL string) *OrderService {
	return &OrderService{Host: host, Port: port, Logger: logger, Session: session, ProductsURL: productsURL}
}

func (svc *OrderService) GetOrdersInTransaction() ([]Order, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")

	var orders []Order
	err := c.Find(bson.M{}).All(&orders)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Find", "Info", err)
		return orders, errors.New(generalErrorInfo)
	}

	return orders, nil
}

func (svc *OrderService) GetOrderByProductId(productId string) (Order, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")

	var order Order
	err := c.Find(bson.M{"product.id": productId}).One(&order)
	if err != nil {
		if err == mgo.ErrNotFound {
			level.Debug(svc.Logger).Log("Product is in order", "false")
			return order, nil
		}

		level.Error(svc.Logger).Log("Find Error", err)
		return order, errors.New(generalErrorInfo)
	}

	return order, nil
}

func (svc *OrderService) Sell(sellInfo Order) (error) {
	level.Debug(svc.Logger).Log("Input", "productId", "value", sellInfo.Product.Id)
	if sellInfo.Product.PriceType == strconv.Itoa(NSUtil.OnlyShow) {
		level.Error(svc.Logger).Log("PriceType", sellInfo.Product.PriceType, "Info", "can't be sold")
		return errors.New("Please change price type!")
	}

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")
	var order Order
	err := c.Find(bson.M{"product.id": sellInfo.Product.Id}).One(&order)
	if err == nil {
		level.Error(svc.Logger).Log("Product", sellInfo.Product.Id, "Info", "is in transaction")
		return errors.New("The product has already been in transaction")
	}

	sellInfo.Status = strconv.Itoa(NSUtil.None)
	sellInfo.ID = NSUtil.UniqueID()
	sellInfo.ServerStartTime = time.Now()
	inputDuration := maxDuration
	if sellInfo.Duration != "" {
		inputDuration,err = strconv.Atoi(sellInfo.Duration)
		if err != nil {
			level.Error(svc.Logger).Log("API", "Atoi", "Error", err)
			return errors.New("Please set right duration")
		}
	}
	if inputDuration > maxDuration {
		sellInfo.Duration = strconv.Itoa(maxDuration)
	}
	err = c.Insert(sellInfo)
	if err != nil {
		level.Error(svc.Logger).Log("Insert error", err)
		return errors.New(generalErrorInfo)
	}

	// sending message to chain
	proTypeString , _ := strconv.Atoi(sellInfo.Product.Type)
	ChainService.StartToSell(sellInfo.ChainId, sellInfo.Product.PriceValue, proTypeString)

	return nil
}

func (svc *OrderService) StopSelling(orderId string) (error) {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderId)
	order,err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Error", err)
		return errors.New(generalErrorInfo)
	}

	return svc.stopSelling(order)
}

func (svc *OrderService) stopSelling(order Order) (error ){
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

	ChainService.StopSelling(order.ChainId)
	return nil
}

func (svc *OrderService) Buy(orderId string, buyInfo BuyInfo) (error) {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderId)
	// get current order
	order,err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Error", err)
		return errors.New(generalErrorInfo)
	}

	// if the product can be bought
	if order.Product.PriceType == strconv.Itoa(NSUtil.Fix) {
		// bought by others 
		if order.Status != strconv.Itoa(NSUtil.None) {
			level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be bought")
			return errors.New("The product has been sold. Please try the others.")
		}
	} else if order.Product.PriceType == strconv.Itoa(NSUtil.Auction) {
		if order.Status != strconv.Itoa(NSUtil.None) && 
		   order.Status != strconv.Itoa(NSUtil.InAuction) {
			level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be bought")
			return errors.New("The product has been sold. Please try the others.")
		}
	} else {
		level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Info", "isn't supported now")
		return errors.New("The product can't be bought. Please try the others.")
	}

	// get current status and send message to chain if necessary
	var updateStatus = strconv.Itoa(NSUtil.None)
	if order.Product.PriceType == strconv.Itoa(NSUtil.Fix) {
		if order.Product.Type == strconv.Itoa(NSUtil.Digit) {
			updateStatus = strconv.Itoa(NSUtil.InFix)

			// send the transaction to chain
			err = ChainService.ConfirmOrder(order.ChainId)
			if err != nil {
				level.Error(svc.Logger).Log("API", "Chain.ConfirmOrder", "Info", err)
				return errors.New(generalErrorInfo)
			}
		} else if order.Product.Type == strconv.Itoa(NSUtil.Entity) {
			updateStatus = strconv.Itoa(NSUtil.Unshipped)
		}
	} else if order.Product.PriceType == strconv.Itoa(NSUtil.Auction) {
		updateStatus = strconv.Itoa(NSUtil.InAuction)

		err = ChainService.UpdatePrice(order.ChainId, buyInfo.Buyer, order.Product.PriceValue)
		if err != nil {
			level.Error(svc.Logger).Log("API", "Chain.UpdatePrice", "Info", err)
			return errors.New(generalErrorInfo)
		}
	} else {
		level.Error(svc.Logger).Log("PriceType", "Chain.UpdatePrice", "Info", err)
		return errors.New("The product can't be bought. Please try the others.")
	}

	// update order info
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")
	buyInfo.ServerStartTime = time.Now()
	updateData := bson.M{"buyinfo": buyInfo,
						 "status": updateStatus}
	err = c.Update(bson.M{"id": orderId}, bson.M{"$set": updateData})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Date.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}


	if testDev {
		if order.Product.PriceType == strconv.Itoa(NSUtil.Fix) {
			svc.ApplyConfirmFromChain(orderId, "success");
		}
	}

	return nil
}

// transaction is successful
func (svc *OrderService) ApplyConfirmFromChain(chainId string, result string) (error) {
	// get order
	order, err := svc.getOrderById(chainId)
	//order, err := svc.getOrderByChainId(chainId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderByChainId", "Info", err)
		return errors.New(generalErrorInfo)
	}

	// if the order can be completed
	if order.Product.PriceType == strconv.Itoa(NSUtil.Fix) {
		if order.Product.Type == strconv.Itoa(NSUtil.Digit) {
			if order.Status != strconv.Itoa(NSUtil.InFix) {
				level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Status", order.Status, "Info", "can't be completed")
				return errors.New("Current order can't be completed")
			}
		} else if order.Product.Type == strconv.Itoa(NSUtil.Entity) {
			if order.Status != strconv.Itoa(NSUtil.DispatchConfirmed) {
				level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Status", order.Status, "Info", "can't be completed")
				return errors.New("Current order can't be completed")
			}
		} else {
			level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Info", "unsupported")
			return errors.New("Current order isn't in a transaction")
		}
	} else if order.Product.PriceType == strconv.Itoa(NSUtil.Auction) {
		if order.Product.Type == strconv.Itoa(NSUtil.Digit) {
			if order.Status != strconv.Itoa(NSUtil.InAuction) {
				level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Status", order.Status, "Info", "can't be completed")
				return errors.New("Current order can't be completed")
			}
		} else if order.Product.Type == strconv.Itoa(NSUtil.Entity) {
			if order.Status != strconv.Itoa(NSUtil.DispatchConfirmed) {
				level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Status", order.Status, "Info", "can't be completed")
				return errors.New("Current order can't be completed")
			}
		} else {
			level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Info", "unsupported")
			return errors.New("Current order isn't in a transaction")
		}
	} else {
		level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Info", "unsupported")
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
	if (err != nil) {
		level.Error(svc.Logger).Log("API", "closeOrder", "Info", err)
		return errors.New(generalErrorInfo)
	} else {
		// update product owner
		if result == "success" {
			svc.updateProductAfterTransaction(order.Product.Id, order.BuyInfo.Buyer, order.BuyInfo.PriceValue)
		}
	}

	return nil
}

func (svc *OrderService) getOrderByChainId(chainId string) (Order, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")

	var order Order
	err := c.Find(bson.M{"chainid": chainId}).One(&order)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Date.find", "Info", err)
		return order, errors.New("can't get order for the chain id " + chainId)
	}

	return order, nil
}

func (svc *OrderService) OrderIsDue(orderId string) (error) {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderId)
	order, err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	// for fix price
	if order.Product.PriceType == strconv.Itoa(NSUtil.Fix) {
		return svc.stopSelling(order)
	} else if order.Product.PriceType == strconv.Itoa(NSUtil.Auction) {
		return svc.auctionIsDue(order)
	} else {
		level.Error(svc.Logger).Log("PriceType", order.Product.PriceType, "Info", "Unsupported")
		return errors.New("The product isn't in transaction")
	}

	return nil
}

func (svc *OrderService) auctionIsDue(order Order) (error) {
	if order.Product.Type == strconv.Itoa(NSUtil.Digit) {
		return ChainService.ConfirmOrder(order.ChainId)
	} else if order.Product.Type == strconv.Itoa(NSUtil.Entity) {
		err := svc.updateOrderStatus(order.ID, NSUtil.Unshipped);
		if err != nil {
			level.Error(svc.Logger).Log("API", "updateOrderStatus", "Info", err)
			return errors.New("Failed to stop auction")
		}
	} else {
		level.Error(svc.Logger).Log("ProductType", order.Product.Type, "Info", "Unsupported")
		return errors.New("unsupported product type for the order " + order.ID)
	}

	return nil
}

func (svc *OrderService) ShipProduct(orderId string, express Express) (error) {
	order, err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.Unshipped) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be shipped")
		return errors.New("Current operation isn't supported. Please check order's status.")
	}

	session := svc.Session.Copy()
	defer session.Close()

	express.StartTime = time.Now()
	updateData := bson.M{"express": express, "status": strconv.Itoa(NSUtil.Dispatched)}
	c := session.DB("store").C("orders")
	err = c.Update(bson.M{"id": orderId}, bson.M{"$set": updateData})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

func (svc *OrderService) getExpressUrl(express Express) (string) {
	// TODO: get link for the express
	var tempValue = express.Company + ":" + express.Number
	return tempValue
}

func (svc *OrderService) ConfirmOrder(orderId string) (error) {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderId)
	order, err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.Dispatched) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be confirmed")
		return errors.New("Current operation isn't supported. Please check order's status.")
	}

	err = svc.updateOrderStatus(orderId, NSUtil.DispatchConfirmed);
	if err != nil {
		level.Error(svc.Logger).Log("API", "updateOrderStatus", "Info", err)
		return errors.New(generalErrorInfo)
	}

	ChainService.ConfirmOrder(order.ChainId)

	if testDev {
		if order.Product.PriceType == strconv.Itoa(NSUtil.Fix) {
			svc.ApplyConfirmFromChain(orderId, "success");
		}
	}
	
	return nil
}

func (svc *OrderService) AskForReturn(orderId string, returnInfo ReturnInfo) (error) {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderId)
	order, err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.Dispatched) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be returned")
		return errors.New("Current operation isn't supported. Please check order's status.")
	}

	session := svc.Session.Copy()
	defer session.Close()

	askStartTime := time.Now()
	savedReturnInfo := svc.convertReturnInfo(order.BuyInfo.Buyer, returnInfo)
	savedReturnInfo.AskTime = askStartTime
	updateData := bson.M{"returninfo": savedReturnInfo, "status": strconv.Itoa(NSUtil.ReturnInAgree)}
	c := session.DB("store").C("orders")
	err = c.Update(bson.M{"id": orderId}, bson.M{"$set": updateData})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

// save returninfo's image from data to a file 
func (svc *OrderService) convertReturnInfo(buyer string, returnInfo ReturnInfo) (ReturnInfo) {
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

func (svc *OrderService) AgreeReturn(orderId string) (error) {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderId)
	order, err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.ReturnInAgree) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be agreed")
		return errors.New("Current operation isn't supported. Please check order's status.")
	}

	session := svc.Session.Copy()
	defer session.Close()

	returnInfo := order.ReturnInfo
	returnInfo.AgreeTime = time.Now()
	updateData := bson.M{"returninfo": returnInfo, "status": strconv.Itoa(NSUtil.ReturnAgreed)}
	c := session.DB("store").C("orders")
	err = c.Update(bson.M{"id": orderId}, bson.M{"$set": updateData})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

func (svc *OrderService) ShipReturn(orderId string, express Express) (error) {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderId)
	order, err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.ReturnAgreed) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be shipped")
		return errors.New("Current operation isn't supported. Please check order's status.")
	}

	session := svc.Session.Copy()
	defer session.Close()

	returnInfo := order.ReturnInfo
	returnInfo.Express = express
	returnInfo.Express.StartTime = time.Now()
	updateData := bson.M{"returninfo": returnInfo, "status": strconv.Itoa(NSUtil.ReturnDispatched)}
	c := session.DB("store").C("orders")
	err = c.Update(bson.M{"id": orderId}, bson.M{"$set": updateData})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

func (svc *OrderService) ConfirmReturn(orderId string) (error) {
	level.Debug(svc.Logger).Log("Input", "orderId", "Value", orderId)
	order, err := svc.getOrderById(orderId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderById", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.ReturnDispatched) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be confirmed")
		return errors.New("Current operation isn't supported. Please check order's status.")
	}

	ChainService.CancelOrder(order.ChainId)

	// update data in database
	session := svc.Session.Copy()
	defer session.Close()

	returnInfo := order.ReturnInfo
	returnInfo.ConfirmTime = time.Now()
	updateData := bson.M{"returninfo": returnInfo, "status": strconv.Itoa(NSUtil.ReturnConfirmed)}
	c := session.DB("store").C("orders")
	err = c.Update(bson.M{"id": orderId}, bson.M{"$set": updateData})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if testDev {
		if order.Product.PriceType == strconv.Itoa(NSUtil.Fix) {
			svc.ApplyCancelFromChain(orderId, "success");
		}
	}

	return nil
}

func (svc *OrderService) ApplyCancelFromChain(chainId string, result string) (error) {
	level.Debug(svc.Logger).Log("Input", "chainId", "Value", chainId)
	order, err := svc.getOrderById(chainId)
	//order, err := svc.getOrderByChainId(chainId)
	if err != nil {
		level.Error(svc.Logger).Log("API", "getOrderByChainId", "Info", err)
		return errors.New(generalErrorInfo)
	}

	if order.Status != strconv.Itoa(NSUtil.ReturnConfirmed) {
		level.Error(svc.Logger).Log("Status", order.Status, "Info", "can't be cancelled")
		return errors.New("Current operation isn't supported. Please check order's status.")
	}

	if result != "success" {
		return errors.New("unhandled return value from chain")
	}

	order.Status = strconv.Itoa(NSUtil.ReturnCompleted)
	err = svc.closeOrder(order)
	if (err != nil) {
		level.Error(svc.Logger).Log("API", "closeOrder", "Info", err)
		return errors.New(generalErrorInfo)
	}

	return nil
}

func (svc *OrderService) updateOrderStatus(orderId string, orderStatus int) (error) {
	session := svc.Session.Copy()
	defer session.Close()

	var currentStatus = strconv.Itoa(orderStatus);
	c := session.DB("store").C("orders")
	err := c.Update(bson.M{"id": orderId}, bson.M{"$set": bson.M{"status": currentStatus}})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Update", "Info", err)
		return errors.New("Failed to complete order")
	}

	return nil
}

// move the order to closed collection
func (svc *OrderService) closeOrder(order Order) (error) {
	level.Debug(svc.Logger).Log("func", "closeOrder")
	session := svc.Session.Copy()
	defer session.Close()

	order.CompleteTime = time.Now()

	c := session.DB("store").C("closedorders")
	err := c.Insert(order)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Insert", "Info", err)
		return errors.New("Failed to close the order")
	} else {
		err = svc.deleteOrder(order.ID);
		if err != nil {
			return errors.New("Failed to delete completed order")
		}
	}

	return nil
}

func (svc *OrderService) GetOrders(buyer string) ([]Order, error) {
	level.Debug(svc.Logger).Log("Input", "buyer", "Value", buyer)
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")

	var orders []Order
	err := c.Find(bson.M{"buyinfo.buyer": buyer}).All(&orders)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Find", "Info", err)
		return orders, errors.New(generalErrorInfo)
	}

	return orders, nil
}

func (svc *OrderService) GetSellings(seller string) ([]Order, error) {
	level.Debug(svc.Logger).Log("Input", "seller", "Value", seller)
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")

	var orders []Order
	err := c.Find(bson.M{"product.owner": seller}).All(&orders)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Find", "Info", err)
		return orders, errors.New(generalErrorInfo)
	}

	return orders, nil
}

func (svc *OrderService) getOrderById(orderId string) (Order, error) {
	level.Debug(svc.Logger).Log("Func", "getOrderById")
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")

	var order Order
	err := c.Find(bson.M{"id": orderId}).One(&order)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Find", "Info", err)
		return order, errors.New("Failed to get order")
	}

	return order, nil
}

func (svc *OrderService) deleteOrder(orderId string) (error) {
	level.Debug(svc.Logger).Log("Func", "deleteOrder")
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("orders")

	err := c.Remove(bson.M{"id": orderId})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Data.Remove", "Info", err)
		return errors.New("Failed to delete order")
	}

	return nil
}

func (svc *OrderService) updateProductAfterTransaction(productId string, newOwner string, price string) (error) {
	updateData := NSUtil.TransactionUpdateData{}
	updateData.Owner = newOwner
	updateData.Price = price
	updateString, _ := json.Marshal(updateData)

	updateClient := &http.Client{}
	updateURL := svc.ProductsURL + "/" + productId + "/transactionupdate"
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