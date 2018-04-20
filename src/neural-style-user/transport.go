package UserService

import (
	"context"
	"encoding/json"

	"net/http"

	"neural-style-util"

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

// MakeHTTPHandler generate the http handler for the style service handler
func MakeHTTPHandler(ctx context.Context, r *mux.Router, svc Service, options ...httptransport.ServerOption) *mux.Router {
	// Register
	registerHandler := httptransport.NewServer(
		MakeNSRegisterEndpoint(svc),
		decodeNSRegisterRequest,
		encodeNSRegisterResponse,
		options...,
	)
	r.Methods("POST").Path("/api/register").Handler(NSUtil.AccessControl(registerHandler))

	// Login
	loginHandler := httptransport.NewServer(
		MakeNSLoginEndpoint(svc),
		decodeNSLoginRequest,
		encodeNSLoginResponse,
		options...,
	)
	r.Methods("POST").Path("/api/authenticate").Handler(NSUtil.AccessControl(loginHandler))

	return r
}
