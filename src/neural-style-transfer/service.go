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
	"time"

	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"io/ioutil"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/dgrijalva/jwt-go"
)

const (
	SecretKey = "Tulian is great"
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

type UserInfo struct {
	ID                string   `json:"id"`
	Name              string   `json:"username"`
	Password          string   `json:"password"`
	Phone             string   `json:"phone"`
	Email             string   `json:"email"`
	ConcernedProducts []string `json:"concernedProducts"`
	ConcernedUsers    []string `json:"concernedUsers"`
}

type UserToken struct {
	ID    string `json:"id"`
	Name  string `json:"username"`
	Token string `json:"token"`
}

// Artist define the basic artist information
type Artist struct {
	Name        string `json:"name"`
	Masterpiece string `json:"masterpiece"`
	ModelName   string `json:"modelname"`
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
	Register(userData UserInfo) (string, error)
	Login(loginData UserInfo) (UserToken, error)
	GetArtists() ([]Artist, error)
	GetHotestArtists() ([]Artist, error)
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
	if strings.HasPrefix(picData, "http") {
		return picData, nil
	}

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
		fmt.Println(err)
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
		fmt.Println(err)
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
		fmt.Println(err)
		return reviews, errors.New("Database error")
	}

	if len(reviews) != 0 {
		return reviews, nil
	}

	return nil, nil
}

// Register create a new user
func (svc NeuralTransferService) Register(userData UserInfo) (string, error) {
	session := svc.Session.Copy()
	defer session.Close()

	// if the user name exists
	var currentUser UserInfo
	c := session.DB("store").C("users")
	err := c.Find(bson.M{"name": userData.Name}).One(&currentUser)
	if err == nil {
		return "", errors.New("User with this name already exists")
	}

	userData.ID = UniqueID()
	result := "Success"
	err = c.Insert(userData)
	if err != nil {
		result = "fail"
		return result, errors.New("Failed to add a new user")
	}

	fmt.Println("Register data is")
	fmt.Println(userData)
	return result, err
}

// Login login the style transfer platform
func (svc NeuralTransferService) Login(loginData UserInfo) (UserToken, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("users")

	var user UserInfo
	err := c.Find(bson.M{"name": loginData.Name}).One(&user)
	if err != nil {
		return UserToken{}, errors.New("This user name is wrong")
	}

	if user.Password != loginData.Password {
		return UserToken{}, errors.New("This password is wrong")
	}

	var userToken UserToken
	userToken.Name = user.Name
	userToken.ID = user.ID
	userToken.Token = CreateToken(loginData.Name)
	return userToken, err
}

// CreateToken create time-limited token
func CreateToken(userName string) string {
	claims := make(jwt.MapClaims)
	claims["username"] = userName
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix() //72小时有效期，过期需要重新登录获取token
	claims["iat"] = time.Now().Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		fmt.Println("Error for sign token: ")
		fmt.Println(err)
		return ""
	}

	return tokenString
}

// CheckToken validate the token
func CheckToken(authString string) (bool, string) {
	authList := strings.Split(authString, " ")
	if len(authList) != 2 || authList[0] != "Bearer" {
		fmt.Println("No authorization info")
		return false, ""
	}

	tokenString := authList[1]
	token, err := jwt.Parse(tokenString, func(*jwt.Token) (interface{}, error) {
		return []byte(SecretKey), nil
	})
	if err != nil {
		fmt.Println("parse claims failed: ", err)
		return false, ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("Can't access claims")
		return false, ""
	}
	user := claims["username"].(string)

	return true, user
}

// GetArtists return all the available artists
func (svc NeuralTransferService) GetArtists() ([]Artist, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("artists")

	var artists []Artist
	err := c.Find(bson.M{}).All(&artists)
	if err != nil {
		// Add log information here
		fmt.Println(err)
		return artists, errors.New("Database error")
	}

	return artists, nil
}

// GetHotestArtists return the active hotest artist
func (svc NeuralTransferService) GetHotestArtists() ([]Artist, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("artists")

	var artists []Artist
	err := c.Find(bson.M{}).All(&artists)
	if err != nil {
		// Add log information here
		fmt.Println(err)
		return artists, errors.New("Database error")
	}

	return artists, nil
}
