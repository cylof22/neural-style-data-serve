package StyleService

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// Service for neural style transfer service
type Service interface {
	StyleTransfer(content, style string, iterations int) (string, error)
	StyleTransferPreview(content, style string) (string, error)
}

// NeuralTransferService for final image style transfer
type NeuralTransferService struct {
	NetworkPath        string
	PreviewNetworkPath string
	OutputPath         string
}

// StyleTransfer for applying the style image to the content image, and generated it as output image
func (svc NeuralTransferService) StyleTransfer(content, style string, iterations int) (string, error) {
	python, err := exec.LookPath("python")
	if err != nil {
		return "", errors.New("No path installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style

	contentPathSepIndex := strings.LastIndex(content, "/")
	contentExtSepIndex := strings.LastIndex(content, ".")
	contentName := content[contentPathSepIndex+1 : contentExtSepIndex-1]

	stylePathSepIndex := strings.LastIndex(style, "/")
	styleExtSepIndex := strings.LastIndex(style, ".")
	styleName := style[stylePathSepIndex+1 : styleExtSepIndex-1]

	output := svc.OutputPath + contentName + "_" + styleName + ".png"
	outputEnv := "output=" + output

	iterationsEnv := "iterations=" + strconv.Itoa(iterations)
	networkPathEnv := "network=" + svc.NetworkPath + "imagenet-vgg-verydeep-19.mat"

	wd, _ := os.Getwd()
	pyfiles := wd + "/neural_style.py"
	cmd := exec.Command(python, pyfiles)
	cmd.Env = []string{targetEnv, styleEnv, outputEnv, iterationsEnv, networkPathEnv}

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil {
		return "", errors.New(fmt.Sprint(err) + ":" + stderr.String())
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		return "", errors.New("Style Transfer fails")
	}
	return output, nil
}

// StyleTransferPreview for applying the style image to the content image, and generated it as output image
func (svc NeuralTransferService) StyleTransferPreview(content, style string) (string, error) {
	python, err := exec.LookPath("python")
	if err != nil {
		return "", errors.New("No path installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style

	contentSepIndex := strings.LastIndex(content, ".")
	styleSepIndex := strings.LastIndex(style, ".")

	contentName := content[0 : contentSepIndex-1]
	styleName := style[0 : styleSepIndex-1]
	output := svc.OutputPath + contentName + "_" + styleName + "_" + "preview" + ".png"
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

	if err != nil {
		return "", errors.New(fmt.Sprint(err) + ":" + stderr.String())
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		return "", errors.New("Style Transfer fails")
	}

	return output, nil
}
