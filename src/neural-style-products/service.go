package ProductService

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/color"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"neural-style-image-watermark"
	"neural-style-util"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// ProductStory define the background story of the image
type ProductStory struct {
	Description string   `json:"description"`
	Pictures    []string `json:"pictures"`
}

// price type
// const (
// 	fix = iota
//     auction
//     onlyShow
// )
// ProductPrice define the basic price and type of the image
type ProductPrice struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// product type
// const (
//     Digit = iota
//     Entity
// )
// UploadProduct define the full information of the uploaded image product
type UploadProduct struct {
	Owner       string       `json:"owner"`
	Maker       string       `json:"maker"`
	Price       ProductPrice `json:"price"`
	PicData     string       `json:"picData"`
	StyleImgURL string       `json:"styleImageUrl"`
	Tags        []string     `json:"tags"`
	Story       ProductStory `json:"story"`
	Type        string       `json:"type"`
}

// BatchProducts define the products information for the batched uploaded image products
type BatchProducts struct {
	Owner    string       `json:"owner"`
	Maker    string       `json:"maker"`
	Price    ProductPrice `json:"price"`
	PicDatas []string     `json:"datas"`
	Tags     []string     `json:"tags"`
	Type     string       `json:"type"`
}

// Product define the basic elements of the product
type Product struct {
	ID          string       `json:"id"`
	Owner       string       `json:"owner"`
	Maker       string       `json:"maker"`
	Price       ProductPrice `json:"price"`
	Rating      float32      `json:"rating"`
	URL         string       `json:"url"`
	StyleImgURL string       `json:"styleImgUrl"`
	Tags        []string     `json:"tags"`
	Story       ProductStory `json:"story"`
	Type        string       `json:"type"`
}

type QueryParams struct {
	Categories []string `json:"categories"`
	Owner      []string `json:"owner"`
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

// ProductService for final image style transfer
type ProductService struct {
	OutputPath   string
	Host         string
	Port         string
	Session      *mgo.Session
	SaveURL      string
	FindURL      string
	CacheSaveURL string
	CacheGetURL  string
}

// NewProductSVC create a new product service
func NewProductSVC(outputPath, host, port, saveURL, findURL, cacheSaveURL, cacheGetURL string, session *mgo.Session) *ProductService {
	return &ProductService{OutputPath: outputPath, Host: host, Port: port, Session: session,
		SaveURL: saveURL, FindURL: findURL, CacheSaveURL: cacheSaveURL, CacheGetURL: cacheGetURL}
}

// CompareExpireTimeinSASWithNow compare the generated expire time of a Azure SAS with now
// True for 1 seconds later than now
func CompareExpireTimeinSASWithNow(sasURL string) bool {
	u, err := url.Parse(sasURL)
	if err == nil {
		urlQuery, err := url.ParseQuery(u.RawQuery)
		if err == nil {
			seArrays := urlQuery["se"]
			if len(seArrays) != 0 {
				expireTime, err := time.Parse(time.RFC3339, seArrays[0])
				if err != nil {
					fmt.Println("Time Parse error")
				}
				diff := expireTime.Sub(time.Now())
				return diff.Seconds() > 1
			}
		}
	}

	return false
}

// upload picture file
func (svc *ProductService) uploadPicutre(owner, picData, picID, picFolder string) (string, error) {
	pos := strings.Index(picData, ",")
	realData := picData[pos+1 : len(picData)]
	baseData, err := base64.StdEncoding.DecodeString(realData)
	if err != nil {
		return "", err
	}

	imgReader := bytes.NewReader(baseData)
	img, format, err := image.Decode(imgReader)
	if err != nil {
		return "", err
	}
	outfileName := picID + "." + format

	storageClient := &http.Client{}

	storageURL := svc.SaveURL + "?userid=" + owner + "&imageid=" + outfileName
	bodyReader := bytes.NewReader(baseData)
	storageReq, err := http.NewRequest("POST", storageURL, bodyReader)
	if err != nil {
		fmt.Println("Construct IO reader fails")
		return "", err
	}

	res, err := storageClient.Do(storageReq)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New("Upload fails")
	}

	watermarkSVC := WaterMark.Service{
		SourceImg: img,
		Text:      "tulian",
		TextColor: color.RGBA{0, 0, 255, 255},
		Scale:     5.0,
		Format:    format,
	}

	//markedFile, err := os.Create(outfileName)
	//defer markedFile.Close()

	markedBytes := make([]byte, 0)
	outputBuffers := bytes.NewBuffer(markedBytes)

	_, err = watermarkSVC.CreateWaterMark(outputBuffers)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}

	// add the memecached item
	cacheClient := &http.Client{}
	cacheURL := svc.CacheSaveURL + "/" + owner + outfileName
	cacheReq, err := http.NewRequest("POST", cacheURL, outputBuffers)
	if err != nil {
		fmt.Println(err.Error())
		return "", err
	}
	res, err = cacheClient.Do(cacheReq)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		fmt.Println(res.Status)
		return "", errors.New("cache fails")
	}

	// construct the memcached url
	fmt.Println("The Cached URL is " + svc.CacheGetURL + "/" + owner + outfileName)
	return svc.CacheGetURL + "/" + owner + outfileName, nil
}

// UploadContentFile upload content file to the cloud storage
func (svc *ProductService) UploadContentFile(productData Product) (Product, error) {
	imageID := NSUtil.UniqueID()
	newImageURL, err := svc.uploadPicutre(productData.Owner, productData.URL, imageID, "contents")

	newContent := Product{ID: imageID}
	if err != nil {
		fmt.Println(err)
		return newContent, err
	}

	newContent.URL = newImageURL
	return newContent, nil
}

// UploadStyleFile upload style file to the cloud storage
func (svc *ProductService) UploadStyleFile(productData UploadProduct) (Product, error) {
	imageID := NSUtil.UniqueID()
	newImageURL, err := svc.uploadPicutre(productData.Owner, productData.PicData, imageID, "styles")

	// The product's URL is a cached local image url, it will be updated by listening the ImageStoreService
	// UploadResult Channel asychonously
	newProduct := Product{ID: imageID}
	if err != nil {
		fmt.Println(err)
		return newProduct, err
	}

	newProduct.Owner = productData.Owner
	newProduct.Maker = productData.Maker
	newProduct.Price = productData.Price
	newProduct.Tags = productData.Tags
	newProduct.URL = newImageURL
	newProduct.StyleImgURL = productData.StyleImgURL
	newProduct.Story.Description = productData.Story.Description
	newProduct.Type = productData.Type

	newProduct.Story.Pictures = productData.Story.Pictures
	for index, pic := range productData.Story.Pictures {
		picId := NSUtil.UniqueID()
		picURL, err := svc.uploadPicutre(productData.Owner, pic, picId, "styles")
		if err == nil {
			newProduct.Story.Pictures[index] = picURL
		} else {
			newProduct.Story.Pictures[index] = ""
		}
	}

	// add it to product data to the database
	err = svc.addProduct(newProduct)

	return newProduct, nil
}

// UploadStyleFiles upload style file
func (svc *ProductService) UploadStyleFiles(products BatchProducts) (string, error) {
	for index, picData := range products.PicDatas {
		var uploadData UploadProduct
		uploadData.Owner = products.Owner
		uploadData.Maker = products.Maker
		uploadData.Price = products.Price
		uploadData.Tags = products.Tags
		uploadData.Type = products.Type
		uploadData.PicData = picData

		_, err := svc.UploadStyleFile(uploadData)
		if err != nil {
			fmt.Println(err)
			return "Number" + strconv.Itoa(index) + " is failed to upload", err
		}
	}

	return "succeed", nil
}

func (svc *ProductService) addProduct(product Product) error {
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

func getQueryBSon(params QueryParams) bson.M {
	query := bson.M{}
	if params.Categories != nil {
		query["categories"] = params.Categories
	}

	if params.Owner != nil {
		query["owner"] = params.Owner[0]
	}

	return query
}

// GetProducts find all the generated products(images)
func (svc *ProductService) GetProducts(params QueryParams) ([]Product, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")

	var products []Product
	query := getQueryBSon(params)
	err := c.Find(query).All(&products)
	if err != nil {
		// Add log information here
		fmt.Println(err)
		return products, errors.New("Database error")
	}

	// get the memcached image data, if cache missing, get the url from the backend storage,
	// watermark the file, and add to the memcache

	return products, nil
}

// GetProductsByID find the product by id
func (svc *ProductService) GetProductsByID(id string) (Product, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")
	var product Product
	err := c.Find(bson.M{"id": id}).One(&product)
	if err != nil {
		fmt.Println(err)
		return Product{}, errors.New("Database error")
	}

	// get the memcached image data, if cache missing, get the url from the backend storage,
	// watermark the file, and add to the memcache

	if product.ID != id {
		return Product{}, errors.New("Failed to find product for the id: " + id)
	}

	return product, nil
}

// GetReviewsByProductID find the
func (svc *ProductService) GetReviewsByProductID(id string) ([]Review, error) {

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
func (svc *ProductService) GetArtists() ([]Artist, error) {
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
func (svc *ProductService) GetHotestArtists() ([]Artist, error) {
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
