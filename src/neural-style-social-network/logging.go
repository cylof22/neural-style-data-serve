package main

import (
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type loggingService struct {
	logger        log.Logger
	socialService Service
}

func newLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (svc *loggingService) GetReviewsByProductID(id string) (reviews []Review, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetReviewsByProductID", "id", id,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.GetReviewsByProductID(id)
}

func (svc *loggingService) AddReviewByProductID(review Review) (err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "AddReviewByProductID", "user", review.User,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.AddReviewByProductID(review)
}

func (svc *loggingService) GetFolloweesByProductID(id string) (followees []Followee, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetFolloweesByProductID", "productid", id, "took",
			time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.GetFolloweesByProductID(id)
}

func (svc *loggingService) AddFolloweesByProductID(user Followee) (err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "AddFolloweesByProductID", "productid", user.ProductID, "user", user.User,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.AddFolloweesByProductID(user)
}

func (svc *loggingService) DeleteFolloweeByID(productID, UserID string) (err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "DeleteFolloweeByProductID", "productid", productID, "userid", UserID,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.DeleteFolloweeByID(productID, UserID)
}

func (svc *loggingService) GetSummaryByID(productID string) (summary SocialSummary, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetSummaryByID", "productid", productID,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.GetSummaryByID(productID)
}

func (svc *loggingService) GetFollowingProductsByUserID(user string) (prods []FollowingProduct, err error) {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "GetFollowingProductsByUserID", "user", user,
			"took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.socialService.GetFollowingProductsByUserID(user)
}

func (svc *loggingService) HealthCheck() bool {
	defer func(begin time.Time) {
		level.Debug(svc.logger).Log("method", "HealthCheck", "took", time.Since(begin))
	}(time.Now())

	return svc.socialService.HealthCheck()
}
