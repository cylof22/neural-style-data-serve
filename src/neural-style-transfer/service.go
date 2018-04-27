package StyleService

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
)

// NeuralTransferService for final image style transfer
type NeuralTransferService struct {
	NetworkPath        string
	PreviewNetworkPath string
	OutputPath         string
	Host               string
	Port               string
}

// NewNeuralTransferSVC generate a transfer service
func NewNeuralTransferSVC(networkPath, previewNetworkPath, outputPath, host, port string) *NeuralTransferService {
	return &NeuralTransferService{NetworkPath: networkPath, PreviewNetworkPath: previewNetworkPath,
		OutputPath: outputPath, Host: host, Port: port}
}

// StyleTransfer for applying the style image to the content image, and generated it as output image
func (svc *NeuralTransferService) StyleTransfer(content, style string, iterations int) (string, error) {
	python, err := exec.LookPath("python")
	if err != nil {
		return "", errors.New("No python installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style

	_, contentName := path.Split(content)
	_, styleName := path.Split(style)

	outputName := contentName + "_" + styleName + ".png"
	output := svc.OutputPath + "data/outputs/" + outputName
	outputEnv := "output=" + output

	iterationsEnv := "iterations=" + strconv.Itoa(iterations)
	networkPathEnv := "network=" + svc.NetworkPath + "imagenet-vgg-verydeep-19.mat"

	fmt.Println("The content path is " + content)
	fmt.Println("The style path is " + style)

	wd, _ := os.Getwd()
	pyfiles := wd + "/neural_style.py"
	cmd := exec.Command(python, pyfiles)
	cmd.Env = []string{targetEnv, styleEnv, outputEnv, iterationsEnv, networkPathEnv}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()

	if _, err := os.Stat(output); os.IsNotExist(err) {
		return "", errors.New("Style Transfer fails")
	}
	return svc.Host + ":" + svc.Port + "/outputs/" + outputName, nil
}

// StyleTransferPreview for applying the style image to the content image, and generated it as output image
func (svc *NeuralTransferService) StyleTransferPreview(content, style string) (string, error) {
	python, err := exec.LookPath("python")
	if err != nil {
		return "", errors.New("No python installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style

	_, contentName := path.Split(content)
	_, styleName := path.Split(style)

	outputName := contentName + "_" + styleName + "_" + "preview" + ".png"
	output := svc.OutputPath + "data/outputs/" + outputName
	outputEnv := "output=" + output

	wd, _ := os.Getwd()
	pyfiles := wd + "/neural_style_preview.py"
	cmd := exec.Command(python, pyfiles)
	cmd.Env = []string{targetEnv, styleEnv, outputEnv}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()

	if _, err := os.Stat(output); os.IsNotExist(err) {
		return "", errors.New("Style Transfer Preview fails")
	}

	return svc.Host + ":" + svc.Port + "/outputs/" + outputName, nil
}
