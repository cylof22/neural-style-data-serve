package main

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

// NSAuthenticationRequest parameters for register and login
type NSAuthenticationRequest struct {
	UserData UserInfo
}

// NSRegisterResponse returns register result
type NSRegisterResponse struct {
	Result string `json:"result"`
	Err    error  `json:"err"`
}

// NSLoginResponse returns token
type NSLoginResponse struct {
	Target UserToken
	Err    error
}

type NSGetUserInfoRequest struct {
	UserName string
}

type NSGetUserInfoResponse struct {
	Target UserInfo
	Err    error
}

type NSAUpdateUserInfoRequest struct {
	UserData UserInfo
}

// NSRegisterResponse returns register result
type NSUpdateUserInfoResponse struct {
	Portrait string
	Err      error
}

// MakeNSRegisterEndpoint generate the endpoint for new user register
func MakeNSRegisterEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAuthenticationRequest)
		res, err := svc.Register(req.UserData)
		return NSRegisterResponse{Result: res, Err: err}, err
	}
}

// MakeNSLoginEndpoint generate the endpoint for user's login
func MakeNSLoginEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAuthenticationRequest)
		token, err := svc.Login(req.UserData)
		return NSLoginResponse{Target: token, Err: err}, err
	}
}

func MakeNSGetUserInfoEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSGetUserInfoRequest)
		userInfo, err := svc.GetUserInfo(req.UserName)
		return NSGetUserInfoResponse{Target: userInfo, Err: err}, err
	}
}

func MakeNSUpdateUserInfoEndpoint(svc Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(NSAUpdateUserInfoRequest)
		result, err := svc.UpdateUserInfo(req.UserData)
		return NSUpdateUserInfoResponse{Portrait: result, Err: err}, err
	}
}
