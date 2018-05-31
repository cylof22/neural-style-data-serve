package miniprogram

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
)

const (
	//miniURL = "https://www.tulian.17dodo.com"
	transferURL = "http://localhost:5000"
	hostSite    = "http://localhost:8000"
)

// MakeHTTPHandler generate the api for miniprogram
func MakeHTTPHandler(ctx context.Context, r *mux.Router) *mux.Router {
	r.Methods("POST").Path("/mini/content").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		file, headers, err := req.FormFile("file")
		if file == nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		contentFilePath := "./data/contents/" + headers.Filename
		contentFile, err := os.Create(contentFilePath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		defer contentFile.Close()

		_, err = io.Copy(contentFile, file)

		w.Write([]byte(hostSite + "/contents/" + headers.Filename))
	})

	r.Methods("POST").Path("/mini/style").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		file, headers, err := req.FormFile("file")
		if file == nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		contentFilePath := "./data/styles/" + headers.Filename
		contentFile, err := os.Create(contentFilePath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		defer contentFile.Close()

		_, err = io.Copy(contentFile, file)

		w.Write([]byte(hostSite + "/styles/" + headers.Filename))
	})

	r.Methods("GET").Path("/mini/styleTransfer").Queries("content", "{content}", "style", "{style}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		styleTransferURL := transferURL + "/styleTransfer"
		vars := mux.Vars(req)
		content := vars["content"]
		style := vars["style"]

		// base64 encode the content and style
		contentURL, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		contentName := filepath.Base(string(contentURL))

		styleURL, err := base64.StdEncoding.DecodeString(style)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		styleName := filepath.Base(string(styleURL))

		styleTransferURL = styleTransferURL + "?content=" + content + "&style=" + style
		transferReq, err := http.NewRequest("GET", styleTransferURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		transferClient := &http.Client{}
		res, err := transferClient.Do(transferReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		outputFilePath := "./data/outputs/" + styleName + contentName
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		_, err = io.Copy(outputFile, res.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write([]byte(hostSite + "/outputs/" + styleName + contentName))
	})

	r.Methods("GET").Path("/mini/fixedStyle").Queries("content", "{content}", "style", "{style}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		styleTransferURL := transferURL + "/fixedStyle"
		vars := mux.Vars(req)
		content := vars["content"]
		style := vars["style"]

		// decode the content
		contentURL, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fileName := filepath.Base(string(contentURL))

		styleTransferURL = styleTransferURL + "?content=" + content + "&style=" + style

		transferReq, err := http.NewRequest("GET", styleTransferURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		transferClient := &http.Client{}
		res, err := transferClient.Do(transferReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		outputFilePath := "./data/outputs/" + style + fileName
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		_, err = io.Copy(outputFile, res.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write([]byte(hostSite + "/outputs/" + style + fileName))
	})

	r.Methods("GET").Path("/mini/artistStyle").Queries("content", "{content}", "artist", "{artist}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		artistTransferURL := transferURL + "/artistStyle"
		vars := mux.Vars(req)
		content := vars["content"]
		artist := vars["artist"]

		// base64 encode the content
		artistTransferURL = artistTransferURL + "?content=" + content + "&artist=" + artist
		transferReq, err := http.NewRequest("GET", artistTransferURL, nil)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}

		transferClient := &http.Client{}
		res, err := transferClient.Do(transferReq)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		contentURL, err := base64.StdEncoding.DecodeString(content)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
		fileName := filepath.Base(string(contentURL))

		outputFilePath := "./data/outputs/" + artist + fileName
		outputFile, err := os.Create(outputFilePath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		_, err = io.Copy(outputFile, res.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write([]byte(hostSite + "/outputs/" + artist + fileName))
	})

	return r
}
