package UserService

import (
	"context"
	"encoding/json"

	"net/http"

	"neural-style-util"

	"github.com/go-kit/kit/endpoint"

	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func decodeNSRegisterRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userData := UserInfo{ID: "1"}
	json.NewDecoder(r.Body).Decode(&userData)
	return NSAuthenticationRequest{UserData: userData}, nil
}

func encodeNSRegisterResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	//if e, ok := response.(errorer); ok && e.error() != nil {
	//	encodeError(ctx, e.error(), w)
	//	return nil
	//}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(response)
}

func decodeNSLoginRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userData := UserInfo{ID: "1"}
	json.NewDecoder(r.Body).Decode(&userData)
	return NSAuthenticationRequest{UserData: userData}, nil
}

func encodeNSLoginResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	loginRes := response.(NSLoginResponse)
	if loginRes.Err != nil {
		return loginRes.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(loginRes.Target)
}

func decodeNSGetUserInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	username := vars["username"]

	return NSGetUserInfoRequest{UserName: username}, nil
}

func encodeNSGetUserInfoResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	res := response.(NSGetUserInfoResponse)
	if res.Err != nil {
		return res.Err
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(res.Target)
}

func decodeNSUpdateUserInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	userData := UserInfo{}
	json.NewDecoder(r.Body).Decode(&userData)
	return NSAUpdateUserInfoRequest{UserData: userData}, nil
}

func encodeNSUpdateUserInfoResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	updateError := response.(NSUpdateUserInfoResponse)
	if updateError.Err != nil {
		return updateError.Err
	}

	return nil
}

// MakeHTTPHandler generate the http handler for the style service handler
func MakeHTTPHandler(ctx context.Context, r *mux.Router, auth endpoint.Middleware, svc Service, options ...httptransport.ServerOption) *mux.Router {
	// Register
	registerHandler := httptransport.NewServer(
		MakeNSRegisterEndpoint(svc),
		decodeNSRegisterRequest,
		encodeNSRegisterResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/register").Handler(NSUtil.AccessControl(registerHandler))

	// Login
	loginHandler := httptransport.NewServer(
		MakeNSLoginEndpoint(svc),
		decodeNSLoginRequest,
		encodeNSLoginResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/authenticate").Handler(NSUtil.AccessControl(loginHandler))

	// GET /api/v1/users/{username}
	r.Methods("GET").Path("/api/v1/users/{username}").Handler(httptransport.NewServer(
		auth(MakeNSGetUserInfoEndpoint(svc)),
		decodeNSGetUserInfoRequest,
		encodeNSGetUserInfoResponse,
		options...,
	))

	updateUserInfoHandler := httptransport.NewServer(
		auth(MakeNSUpdateUserInfoEndpoint(svc)),
		decodeNSUpdateUserInfoRequest,
		encodeNSUpdateUserInfoResponse,
		options...,
	)
	r.Methods("POST").Path("/api/v1/users/{username}/update").Handler(NSUtil.AccessControl(updateUserInfoHandler))

	return r
}
