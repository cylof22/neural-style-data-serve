package StyleService

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

// Service for neural style transfer service
type Service interface {
	StyleTransfer(content, style, output string, iterations int) error
	StyleTransferPreview(content, style, output string) error
}

// NeuralTransferService for final image style transfer
type NeuralTransferService struct {
	NetworkPath        string
	PreviewNetworkPath string
}

// StyleTransfer for applying the style image to the content image, and generated it as output image
func (svc NeuralTransferService) StyleTransfer(content, style, output string, iterations int) error {
	python, err := exec.LookPath("python")
	if err != nil {
		return errors.New("No path installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style
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
		return errors.New(fmt.Sprint(err) + ":" + stderr.String())
	}

	return nil
}

// StyleTransferPreview for applying the style image to the content image, and generated it as output image
func (svc NeuralTransferService) StyleTransferPreview(content, style, output string) error {
	python, err := exec.LookPath("python")
	if err != nil {
		return errors.New("No path installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style
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
		return errors.New(fmt.Sprint(err) + ":" + stderr.String())
	}

	return nil
}
