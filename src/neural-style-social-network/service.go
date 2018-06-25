package main

import (
	"errors"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Review define the basic elements of the review
type Review struct {
	ID        uint32 `json:"id"`
	ProductID string `json:"productid"`
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	Rating    uint8  `json:"rating"`
	Comment   string `json:"comment"`
}

// Followee product followee information
type Followee struct {
	ProductID string `json:"productid"`
	User      string `json:"user"`
	Timestamp string `json:"timestamp"`
}

// SocialSummary define the basic social information for the home page
type SocialSummary struct {
	FolloweeCount uint32 `json:"followeeCount"`
	StarRated     uint32 `json:"starRated"`
	CommentCount  uint32 `json:"commentCount"`
}

// FollowingProduct define the following product info
type FollowingProduct struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Service define basic service interface for social network
type Service interface {
	HealthCheck() bool
	GetSummaryByID(id string) (SocialSummary, error)
	GetReviewsByProductID(id string) ([]Review, error)
	AddReviewByProductID(review Review) error
	GetFolloweesByProductID(id string) ([]Followee, error)
	AddFolloweesByProductID(use Followee) error
	DeleteFolloweeByID(productID, User string) error
	GetFollowingProductsByUserID(id string) ([]FollowingProduct, error)
}

// SocialService define implementation of the social service
type SocialService struct {
	Session *mgo.Session
	Logger  log.Logger
}

func newSocialSVC(logger log.Logger, session *mgo.Session) Service {
	return &SocialService{Session: session, Logger: logger}
}

// GetReviewsByProductID find the
func (svc *SocialService) GetReviewsByProductID(id string) ([]Review, error) {

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("reviews")

	var reviews []Review
	err := c.Find(bson.M{"productid": id}).All(&reviews)
	if err != nil {
		// Add log information here
		level.Debug(svc.Logger).Log("API", "GetReviewsByProductID", "info", err.Error(), "id", id)
		return reviews, errors.New("Database error")
	}

	if len(reviews) != 0 {
		return reviews, nil
	}

	level.Debug(svc.Logger).Log("API", "GetReviewsByProductID", "info", "get reviews by id ok", "id", id)
	return nil, nil
}

// AddReviewByProductID add review data to the product id
func (svc *SocialService) AddReviewByProductID(review Review) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("reviews")

	existQuery := c.Find(bson.M{"productid": review.ProductID, "user": review.User}).Limit(1)
	if existQuery != nil {
		reviewCount, err := existQuery.Count()
		if err != nil {
			level.Error(svc.Logger).Log("API", "AddFolloweesByProductID", "user", review.User, "productid", review.ProductID, "info", err.Error())
			return err
		}

		if reviewCount != 0 {
			level.Error(svc.Logger).Log("API", "AddFolloweesByProductID", "user", review.User, "productid", review.ProductID, "info", "Duplicated review")
			return errors.New("Duplicated user: " + review.User + " for " + review.ProductID)
		}
	}
	err := c.Insert(review)

	if err != nil {
		level.Error(svc.Logger).Log("API", "GetReviewsByProductID", "info", err.Error())
	}

	level.Debug(svc.Logger).Log("API", "AddReviewByProductID", "user", review.User, "id", review.ProductID, "comments", review.Comment)
	return err
}

// GetFolloweesByProductID get all the followees for a given product id
func (svc *SocialService) GetFolloweesByProductID(id string) ([]Followee, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("followees")

	var followees []Followee
	err := c.Find(bson.M{"productid": id}).All(&followees)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetFolloweesByProductID", "productid", id, "info", err.Error())
		return nil, err
	}

	if len(followees) != 0 {
		return followees, nil
	}

	level.Debug(svc.Logger).Log("API", "GetFolloweesByProductID", "productid", id, "size", len(followees))
	return followees, nil
}

// AddFolloweesByProductID add new followees to a given product id
func (svc *SocialService) AddFolloweesByProductID(info Followee) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("followees")

	existQuery := c.Find(bson.M{"productid": info.ProductID, "user": info.User}).Limit(1)
	if existQuery != nil {
		followwCount, err := existQuery.Count()
		if err != nil {
			level.Error(svc.Logger).Log("API", "AddFolloweesByProductID", "user", info.User, "productid", info.ProductID, "info", err.Error())
			return err
		}

		if followwCount != 0 {
			level.Error(svc.Logger).Log("API", "AddFolloweesByProductID", "user", info.User, "productid", info.ProductID, "info", "Duplicated followee")
			return errors.New("Duplicated user: " + info.User + " for " + info.ProductID)
		}
	}

	err := c.Insert(info)

	if err != nil {
		level.Error(svc.Logger).Log("API", "AddFolloweesByProductID", "user", info.User, "productid", info.ProductID, "info", err.Error())
		return err
	}

	level.Debug(svc.Logger).Log("API", "AddFolloweesByProductID", "user", info.User, "id", info.ProductID)
	return nil
}

// DeleteFolloweeByID remove the followee information for a given product id
func (svc *SocialService) DeleteFolloweeByID(productID, User string) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("followees")

	err := c.Remove(bson.M{"productid": productID, "user": User})
	if err != nil {
		level.Error(svc.Logger).Log("API", "DeleteFolloweeByIDProductID", "productid", productID, "user", User, "info", err.Error())
		return err
	}

	level.Debug(svc.Logger).Log("API", "DeleteFolloweeByIDProductID", "productid", productID, "user", User)
	return err
}

// GetSummaryByID aggregate the summary information from the 'reviews' and 'followees' collection
func (svc *SocialService) GetSummaryByID(id string) (SocialSummary, error) {
	session := svc.Session.Copy()
	defer session.Close()

	info := SocialSummary{}

	c := session.DB("store").C("followees")
	count, err := c.Find(bson.M{"productid": id}).Count()
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetSummaryByID", "productid", id, "info", "GetFollowees", "error", err.Error())
		return info, err
	}
	info.FolloweeCount = uint32(count)

	c = session.DB("store").C("reviews")
	var ratings []struct {
		Rating uint8 `json:"rating"`
	}

	err = c.Find(bson.M{"productid": id}).Select(bson.M{"rating": 1}).All(&ratings)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetSummaryByID", "productid", id, "info", "GetReviews", "error", err.Error())
		return info, err
	}
	info.CommentCount = uint32(len(ratings))

	totalRating := 0
	for _, ratingInfo := range ratings {
		totalRating += int(ratingInfo.Rating)
	}

	info.StarRated = 0
	if info.CommentCount != 0 {
		info.StarRated = uint32(totalRating) / info.CommentCount
	}

	level.Info(svc.Logger).Log("API", "GetSummaryByID", "productid", id, "followees", info.FolloweeCount, "comments", info.CommentCount, "rating", info.StarRated)
	return info, nil
}

// GetFollowingProductsByUserID get all the following product for a given user name
func (svc *SocialService) GetFollowingProductsByUserID(user string) ([]FollowingProduct, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("followees")

	var prodIds []struct {
		ProductID string `json:"productid"`
	}

	err := c.Find(bson.M{"user": user}).Select(bson.M{"productid": 1}).All(&prodIds)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetFollowingProductsByUserID", "user", user, "info", err.Error())
		return nil, err
	}

	if len(prodIds) == 0 {
		return nil, nil
	}

	var idArray []string
	for _, info := range prodIds {
		idArray = append(idArray, info.ProductID)
	}
	// aggregate all the following product infor from the products collection
	var prods []FollowingProduct
	prodC := session.DB("store").C("products")
	err = prodC.Find(bson.M{"id": bson.M{"$in": idArray}}).Select(bson.M{"id": 1, "url": 1}).All(&prods)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetFollowingProductsByUserID", "user", user, "info", err.Error())
		return nil, err
	}

	level.Info(svc.Logger).Log("API", "GetFollowingProductsByUserID", "user", user, "info", len(prods))
	return prods, nil
}

// HealthCheck return the status of the social service by checking the mongodb, or network
func (svc *SocialService) HealthCheck() bool {
	//
	return true
}
