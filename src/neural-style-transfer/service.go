package StyleService

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

// Product define the basic elements of the product
type Product struct {
	ID          uint64   `json:"id"`
	Title       string   `json:"title"`
	Price       float64  `json:"price"`
	Rating      float32  `json:"rating"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Categories  []string `json:"categories"`
}

// Review define the basic elements of the review
type Review struct {
	ID        uint32 `json:"id"`
	ProductID uint64 `json:"productId"`
	Timestamp string `json:"timestamp"`
	User      string `json:"user"`
	Rating    uint8  `json:"rating"`
	Comment   string `json:"comment"`
}

// Service for neural style transfer service
type Service interface {
	StyleTransfer(content, style string, iterations int) (string, error)
	StyleTransferPreview(content, style string) (string, error)
	UploadContentFile(name string, imgFile multipart.File) (string, error)
	UploadStyleFile(name string, imgFile multipart.File) (string, error)
	GetProducts() ([]Product, error)
	GetProductsByID(id uint64) (Product, error)
	GetReviewsByProductID(id uint64) ([]Review, error)
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
		return "", errors.New("No python installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style

	contentPathSepIndex := strings.LastIndex(content, "/")
	contentExtSepIndex := strings.LastIndex(content, ".")
	contentName := content[contentPathSepIndex+1 : contentExtSepIndex]

	stylePathSepIndex := strings.LastIndex(style, "/")
	styleExtSepIndex := strings.LastIndex(style, ".")
	styleName := style[stylePathSepIndex+1 : styleExtSepIndex]

	output := svc.OutputPath + "data/outputs/" + contentName + "_" + styleName + ".png"
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
	return output, nil
}

// StyleTransferPreview for applying the style image to the content image, and generated it as output image
func (svc NeuralTransferService) StyleTransferPreview(content, style string) (string, error) {
	python, err := exec.LookPath("python")
	if err != nil {
		return "", errors.New("No python installed")
	}

	targetEnv := "content=" + content
	styleEnv := "styles=" + style

	contentPathSepIndex := strings.LastIndex(content, "/")
	contentExtSepIndex := strings.LastIndex(content, ".")
	contentName := content[contentPathSepIndex+1 : contentExtSepIndex]

	stylePathSepIndex := strings.LastIndex(style, "/")
	styleExtSepIndex := strings.LastIndex(style, ".")
	styleName := style[stylePathSepIndex+1 : styleExtSepIndex]

	output := svc.OutputPath + "data/outputs/" + contentName + "_" + styleName + "_" + "preview" + ".png"
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

	return output, nil
}

// UploadContentFile upload content file to the cloud storage
func (svc NeuralTransferService) UploadContentFile(name string, imgFile multipart.File) (string, error) {
	data, err := ioutil.ReadAll(imgFile)
	if err != nil {
		return "", errors.New("Faile to read the Content file")
	}

	outfilename := path.Join("./data/contents/", path.Ext(name))
	err = ioutil.WriteFile(outfilename, data, 0777)
	if err != nil {
		return "", errors.New("Failed to creat the contents file")
	}

	return "", nil
}

// UploadStyleFile upload style file to the cloud storage
func (svc NeuralTransferService) UploadStyleFile(name string, imgFile multipart.File) (string, error) {
	data, err := ioutil.ReadAll(imgFile)
	if err != nil {
		return "", errors.New("Faile to read the Style file")
	}

	outfilename := path.Join("./data/styles/", path.Ext(name))
	err = ioutil.WriteFile(outfilename, data, 0777)
	if err != nil {
		return "", errors.New("Failed to creat the style file")
	}

	return outfilename, nil
}

var allProducts = []Product{
	{
		0,
		"First Product",
		24.99,
		4.3,
		"This is a short description. Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"http://localhost:9090/outputs/16-output.jpg",
		[]string{"electronics", "hardware"}},
	{
		1,
		"Second Product",
		64.99,
		3.5,
		"This is a short description. Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"http://localhost:9090/outputs/11-output.jpg",
		[]string{"books"},
	},
	{
		2,
		"Third Product",
		74.99,
		4.2,
		"This is a short description. Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"http://localhost:9090/outputs/12-output.jpg",
		[]string{"electronics"},
	},
	{
		3,
		"Fourth Product",
		84.99,
		3.9,
		"This is a short description. Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"http://localhost:9090/outputs/13-output.jpg",
		[]string{"hardware"},
	},
	{
		4,
		"Fifth Product",
		94.99,
		5,
		"This is a short description. Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"http://localhost:9090/outputs/14-output.jpg",
		[]string{"electronics", "hardware"},
	},
	{
		5,
		"Sixth Product",
		54.99,
		4.6,
		"This is a short description. Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
		"http://localhost:9090/outputs/15-output.jpg",
		[]string{"books"},
	},
}

// GetProducts find all the generated products(images)
func (svc NeuralTransferService) GetProducts() ([]Product, error) {
	return allProducts, nil
}

// GetProductsByID find the product by id
func (svc NeuralTransferService) GetProductsByID(id uint64) (Product, error) {
	for _, prod := range allProducts {
		if prod.ID == id {
			return prod, nil
		}
	}
	return Product{}, errors.New("No Product for the " + strconv.FormatUint(id, 10))
}

var reviews = []Review{
	{
		0,
		0,
		"2014-05-20T02:17:00+00:00",
		"User 1",
		5,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		1,
		0,
		"2014-05-20T02:53:00+00:00",
		"User 2",
		3,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		2,
		0,
		"2014-05-20T05:26:00+00:00",
		"User 3",
		4,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		3,
		0,
		"2014-05-20T07:20:00+00:00",
		"User 4",
		4,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		4,
		0,
		"2014-05-20T11:35:00+00:00",
		"User 5",
		5,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		5,
		0,
		"2014-05-20T11:42:00+00:00",
		"User 6",
		5,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
}

// GetReviewsByProductID find the
func (svc NeuralTransferService) GetReviewsByProductID(id uint64) ([]Review, error) {
	var selectedReview []Review

	for _, review := range reviews {
		if review.ProductID == id {
			selectedReview = append(selectedReview, review)
		}
	}

	if len(selectedReview) != 0 {
		return selectedReview, nil
	}

	return nil, errors.New("No reviews for target id " + strconv.FormatUint(id, 10))
}
