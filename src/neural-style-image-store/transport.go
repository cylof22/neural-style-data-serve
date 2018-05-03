package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
)

func decodeNSSaveRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	userID := vars["userid"]
	imageID := vars["imageid"]

	imgData, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return NSSaveRequest{UserID: userID, ImageID: imageID, ImageData: imgData}, nil
}

func encodeNSSaveResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	saveRes := response.(NSSaveResponse)

	if saveRes.SaveError != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(saveRes)
}

func decodeNSFindRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	userID := vars["userid"]
	imageID := vars["imageid"]

	return NSFindRequest{UserID: userID, ImageID: imageID}, nil
}

func encodeNSFindResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	findRes := response.(NSFindResponse)

	if findRes.FindError != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("context-type", "application/json, charset=utf8")
	return json.NewEncoder(w).Encode(findRes)
}
