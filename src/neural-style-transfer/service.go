package StyleService

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"

	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// Product define the basic elements of the product
type Product struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Owner       string   `json:"owner"`
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
	UploadContentFile(productData Product) (Product, error)
	UploadStyleFile(productData Product) (Product, error)
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
	Session            *mgo.Session
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

//GetMd5String 生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

//UniqueID 生成Guid字串
func UniqueID() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

// upload picture file
func uploadPicutre(picData string, picID string, picFolder string) (string, error) {
	outfileName := picID + ".png"
	outfilePath := path.Join("./data", picFolder, outfileName)

	pos := strings.Index(picData, ",")
	realData := picData[pos+1 : len(picData)]
	baseData, _ := base64.StdEncoding.DecodeString(realData)
	err := ioutil.WriteFile(outfilePath, baseData, 0644)
	if err != nil {
		return "", err
	}

	newImageURL := "http://localhost:8000/" + picFolder + "/" + outfileName
	fmt.Println("New picuture is created: " + newImageURL)
	return newImageURL, nil
}

// UploadContentFile upload content file to the cloud storage
func (svc NeuralTransferService) UploadContentFile(productData Product) (Product, error) {
	imageID := UniqueID()
	newImageURL, err := uploadPicutre(productData.URL, imageID, "contents")

	newContent := Product{ID: imageID}
	if err != nil {
		fmt.Println(err)
		return newContent, err
	}

	newContent.URL = newImageURL
	return newContent, nil
}

// UploadStyleFile upload style file to the cloud storage
func (svc NeuralTransferService) UploadStyleFile(productData Product) (Product, error) {
	imageID := UniqueID()
	newImageURL, err := uploadPicutre(productData.URL, imageID, "styles")

	newProduct := Product{ID: imageID}
	if err != nil {
		fmt.Println(err)
		return newProduct, err
	}

	newProduct.Owner = productData.Owner
	newProduct.Creator = productData.Creator
	if newProduct.Creator == "" {
		newProduct.Creator = productData.Owner
	}
	newProduct.Title = productData.Title
	newProduct.Description = productData.Description
	newProduct.Price = productData.Price
	newProduct.Categories = productData.Categories
	newProduct.URL = newImageURL
	newProduct.StyleImgURL = productData.StyleImgURL

	// add it to product data to the database
	err = svc.addProduct(newProduct)

	return newProduct, nil
}

func (svc NeuralTransferService) addProduct(product Product) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")

	err := c.Insert(product)
	if err != nil {
		if mgo.IsDup(err) {
			return errors.New("Book with this ISBN already exists")
		}
		return errors.New("Failed to add a new products")
	}

	return nil
}

// GetProducts find all the generated products(images)
func (svc NeuralTransferService) GetProducts() ([]Product, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")

	var products []Product
	err := c.Find(bson.M{}).All(&products)
	if err != nil {
		// Add log information here
		return products, errors.New("Database error")
	}

	return products, nil
}

// GetProductsByID find the product by id
func (svc NeuralTransferService) GetProductsByID(id string) (Product, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")
	var product Product
	err := c.Find(bson.M{"id": id}).One(&product)
	if err != nil {
		return Product{}, errors.New("Database error")
	}

	if product.ID != id {
		return Product{}, errors.New("Failed to find product for the id: " + id)
	}

	return product, nil
}

// GetReviewsByProductID find the
func (svc NeuralTransferService) GetReviewsByProductID(id string) ([]Review, error) {

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("reviews")

	var reviews []Review
	err := c.Find(bson.M{"productId": id}).All(&reviews)
	if err != nil {
		// Add log information here
		return reviews, errors.New("Database error")
	}

	if len(reviews) != 0 {
		return reviews, nil
	}

	return nil, nil
}
