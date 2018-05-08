package UserService

import (
	"time"

	"github.com/go-kit/kit/log"
)

type loggingService struct {
	logger       log.Logger
	loginService Service
}

// NewLoggingService returns a new instance of a products logging Service.
func NewLoggingService(logger log.Logger, s Service) Service {
	return &loggingService{logger, s}
}

func (svc *loggingService) Register(userData UserInfo) (info string, err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "Register", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.loginService.Register(userData)
}

func (svc *loggingService) Login(loginData UserInfo) (token UserToken, err error) {
	defer func(begin time.Time) {
		svc.logger.Log("method", "Login", "took", time.Since(begin), "err", err)
	}(time.Now())

	return svc.loginService.Login(loginData)
}
