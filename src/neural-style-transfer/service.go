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
}

// NeuralTransferService for final image style transfer
type NeuralTransferService struct {
}

// StyleTransfer for applying the style image to the content image, and generated it as output image
func (NeuralTransferService) StyleTransfer(content, style, output string, iterations int) error {
	python, err := exec.LookPath("python")
	if err != nil {
		return errors.New("No path installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style
	outputEnv := "output=" + output
	iterationsEnv := "iterations=" + strconv.Itoa(iterations)

	wd, _ := os.Getwd()
	pyfiles := wd + "/neural_style.py"
	cmd := exec.Command(python, pyfiles)
	cmd.Env = []string{targetEnv, styleEnv, outputEnv, iterationsEnv}

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
