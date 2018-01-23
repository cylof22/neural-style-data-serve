package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
)

func TestNeuralStyleServer(t *testing.T) {
	client := &http.Client{}

	neuralStyleURL := "http://localhost:9090/styleTransfer"

	contentPath := "/Users/caoyuanle/Documents/style-demo/neural-style/examples/1-content.jpg"
	contentQuery := "content=" + base64.StdEncoding.EncodeToString([]byte(contentPath))

	stylePath := "/Users/caoyuanle/Documents/style-demo/neural-style/examples/1-style.jpg"
	styleQuery := "style=" + base64.StdEncoding.EncodeToString([]byte(stylePath))

	outputPath := "/Users/caoyuanle/Documents/style-demo/neural-style/examples/1-output.jpg"
	outputQuery := "output=" + base64.StdEncoding.EncodeToString([]byte(outputPath))

	iterationQuery := "iterations=1"

	neuralStyleURL += "?" + contentQuery + "&" + styleQuery + "&" + outputQuery + "&" + iterationQuery

	fmt.Println(neuralStyleURL)
	req, err := http.NewRequest("GET", neuralStyleURL, nil)
	if err != nil {
		t.Error("Failed to launch a request")
	}

	_, err = client.Do(req)
	if err != nil {
		t.Error("Failed to get the response")
	}
}
