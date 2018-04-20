package ProductService

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"neural-style-util"

	mgo "gopkg.in/mgo.v2"
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

// Artist define the basic artist information
type Artist struct {
	Name        string `json:"name"`
	Masterpiece string `json:"masterpiece"`
	ModelName   string `json:"modelname"`
}

// Service for neural style transfer service
type Service interface {
	UploadContentFile(productData Product) (Product, error)
	UploadStyleFile(productData Product) (Product, error)
	GetProducts() ([]Product, error)
	GetProductsByID(id string) (Product, error)
	GetReviewsByProductID(id string) ([]Review, error)
	GetArtists() ([]Artist, error)
	GetHotestArtists() ([]Artist, error)
}

// ProductService for final image style transfer
type ProductService struct {
	OutputPath string
	Host       string
	Port       string
	Session    *mgo.Session
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
func (svc ProductService) UploadContentFile(productData Product) (Product, error) {
	imageID := NSUtil.UniqueID()
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
func (svc ProductService) UploadStyleFile(productData Product) (Product, error) {
	imageID := NSUtil.UniqueID()
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

func (svc ProductService) addProduct(product Product) error {
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
func (svc ProductService) GetProducts() ([]Product, error) {
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
func (svc ProductService) GetProductsByID(id string) (Product, error) {
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
func (svc ProductService) GetReviewsByProductID(id string) ([]Review, error) {

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

// GetArtists return all the available artists
func (svc ProductService) GetArtists() ([]Artist, error) {
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
func (svc ProductService) GetHotestArtists() ([]Artist, error) {
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