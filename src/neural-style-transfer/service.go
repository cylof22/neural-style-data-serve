package StyleService

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path"
	"strconv"
)

// Product define the basic elements of the product
type Product struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Owener      string   `json:"owner"`
	Creator     string   `json:"creator"`
	Price       float32  `json:"price"`
	Rating      float32  `json:"rating"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	StyleImgURL string   `json:"styleImgUrl"`
	Categories  []string `json:"categories"`
}

// Review define the basic elements of the review
type Review struct {
	ID        uint32 `json:"id"`
	ProductID string `json:"productId"`
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
	GetProductsByID(id string) (Product, error)
	GetReviewsByProductID(id string) ([]Review, error)
}

// NeuralTransferService for final image style transfer
type NeuralTransferService struct {
	NetworkPath        string
	PreviewNetworkPath string
	OutputPath         string
	Host               string
	Port               string
}

// StyleTransfer for applying the style image to the content image, and generated it as output image
func (svc NeuralTransferService) StyleTransfer(content, style string, iterations int) (string, error) {
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
func (svc NeuralTransferService) StyleTransferPreview(content, style string) (string, error) {
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

// UploadContentFile upload content file to the cloud storage
func (svc NeuralTransferService) UploadContentFile(name string, imgFile multipart.File) (string, error) {
	outfilename := path.Join("./data/contents/", name)

	savedFile, err := os.Create(outfilename)
	if err != nil {
		return "", err
	}
	defer savedFile.Close()

	_, err = io.Copy(savedFile, imgFile)
	if err != nil {
		return "", err
	}

	return outfilename, nil
}

// UploadStyleFile upload style file to the cloud storage
func (svc NeuralTransferService) UploadStyleFile(name string, imgFile multipart.File) (string, error) {
	outfilename := path.Join("./data/styles/", name)

	savedFile, err := os.Create(outfilename)
	if err != nil {
		return "", errors.New("Failed to create the style file")
	}

	defer savedFile.Close()

	_, err = io.Copy(savedFile, imgFile)
	if err != nil {
		return "", err
	}

	return outfilename, nil
}

func readProducts() []Product {
	file := "./data/info/images.json"
	var products []Product

	inFile, err := os.Open(file)
	defer inFile.Close()

	if err != nil {
		fmt.Println("Read Products Error" + err.Error())
		return nil
	}

	decoder := json.NewDecoder(inFile)
	err = decoder.Decode(&products)
	if err != nil {
		fmt.Println("decode Products error" + err.Error())
		return nil
	}

	return products
}

// GetProducts find all the generated products(images)
func (svc NeuralTransferService) GetProducts() ([]Product, error) {
	allProducts := readProducts()
	for _, prod := range allProducts {
		fmt.Println(prod.Owener)
	}
	return allProducts, nil
}

// GetProductsByID find the product by id
func (svc NeuralTransferService) GetProductsByID(id string) (Product, error) {
	allProducts := readProducts()
	for _, prod := range allProducts {
		if prod.ID == id {
			return prod, nil
		}
	}
	return Product{}, errors.New("No Product for the " + id)
}

var reviews = []Review{
	{
		0,
		"0",
		"2014-05-20T02:17:00+00:00",
		"User 1",
		5,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		1,
		"0",
		"2014-05-20T02:53:00+00:00",
		"User 2",
		3,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		2,
		"0",
		"2014-05-20T05:26:00+00:00",
		"User 3",
		4,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		3,
		"0",
		"2014-05-20T07:20:00+00:00",
		"User 4",
		4,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		4,
		"0",
		"2014-05-20T11:35:00+00:00",
		"User 5",
		5,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
	{
		5,
		"0",
		"2014-05-20T11:42:00+00:00",
		"User 6",
		5,
		"Aenean vestibulum velit id placerat posuere. Praesent placerat mi ut massa tempor, sed rutrum metus rutrum. Fusce lacinia blandit ligula eu cursus. Proin in lobortis mi. Praesent pellentesque auctor dictum. Nunc volutpat id nibh quis malesuada. Curabitur tincidunt luctus leo, quis condimentum mi aliquet eu. Vivamus eros metus, convallis eget rutrum nec, ultrices quis mauris. Praesent non lectus nec dui venenatis pretium.",
	},
}

// GetReviewsByProductID find the
func (svc NeuralTransferService) GetReviewsByProductID(id string) ([]Review, error) {
	var selectedReview []Review

	for _, review := range reviews {
		if review.ProductID == id {
			selectedReview = append(selectedReview, review)
		}
	}

	if len(selectedReview) != 0 {
		return selectedReview, nil
	}

	return nil, nil
}
