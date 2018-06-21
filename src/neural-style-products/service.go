package ProductService

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"mime"
	"net/http"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/go-kit/kit/log/level"

	"neural-style-image-watermark"
	"neural-style-util"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/go-kit/kit/log"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var memCacheServices = os.Getenv("MEMCACHE_SERVICES")

// ProductStory define the background story of the image
type ProductStory struct {
	Description string   `json:"description"`
	Pictures    []string `json:"pictures"`
}

// ProductPrice define the basic price and type of the image
// price type
// const (
// 	fix = iota
//     auction
//     onlyShow
// )
type ProductPrice struct {
	Type     string `json:"type"`
	Value    string `json:"value"`
	Duration string `json:"duration"`
}

// UploadProduct define the full information of the uploaded image product
// product type
// const (
//     Digit = iota
//     Entity
// )
type UploadProduct struct {
	Owner       string       `json:"owner"`
	Maker       string       `json:"maker"`
	Price       ProductPrice `json:"price"`
	PicData     string       `json:"picData"`
	StyleImgURL string       `json:"styleImageUrl"`
	Tags        []string     `json:"tags"`
	Story       ProductStory `json:"story"`
	Type        string       `json:"type"`
	ChainId     string       `json:"chainId"`
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
	ChainId     string       `json:"chainId"`
}

// Artist define the basic artist information
type Artist struct {
	Name        string `json:"name"`
	Masterpiece string `json:"masterpiece"`
	ModelName   string `json:"modelname"`
}

// Service define the basic interface for the products service
type Service interface {
	// UploadContentFile upload the content data into the service
	UploadContentFile(productData Product) (Product, error)
	UploadStyleFile(productData UploadProduct) (Product, error)
	UploadStyleFiles(products BatchProducts) (string, error)
	GetProducts() ([]Product, error)
	GetProductsByUser(userID string) ([]Product, error)
	GetProductsByTags(tag []string) ([]Product, error)
	GetProductsByID(id string) (Product, error)
	GetArtists() ([]Artist, error)
	GetHotestArtists() ([]Artist, error)
	GetImage(userID, imageID string) ([]byte, string, error)
	DeleteProduct(productID string) error
	UpdateProduct(productID string, productData UploadProduct) error
	UpdateProductAfterTransaction(productId string, newOwner string, newPrice string) error
	Search(keyvals map[string]interface{}) ([]Product, error)
}

// ProductService for final image style transfer
type ProductService struct {
	OutputPath  string
	Host        string
	Port        string
	Session     *mgo.Session
	SaveURL     string
	FindURL     string
	CacheGetURL string
	IsLocalDev  bool
	CacheClient *memcache.Client
	Logger      log.Logger
}

// NewProductSVC create a new product service
func NewProductSVC(outputPath, host, port, saveURL, findURL, cacheGetURL string, localDev bool, logger log.Logger,
	session *mgo.Session) *ProductService {
	var client *memcache.Client
	if !localDev {
		var memcachedURL []string
		if len(memCacheServices) == 0 {
			memcachedURL = append(memcachedURL, "localhost:11211")
		} else {
			memcachedURL = strings.Split(memCacheServices, ";")
		}

		//create a handle
		client = memcache.New(memcachedURL...)
		if client == nil {
			level.Error(logger).Log("memcache", "Failed to connect to the memcache server")
		}
	} else {
		client = nil
	}

	return &ProductService{OutputPath: outputPath, Host: host, Port: port, Session: session,
		SaveURL: saveURL, FindURL: findURL, CacheGetURL: cacheGetURL, IsLocalDev: localDev,
		Logger: logger, CacheClient: client}
}

// upload picture file
func (svc *ProductService) uploadPicture(owner, picData, picID, picFolder string) (string, error) {
	pos := strings.Index(picData, ",")
	if len(picData) < 11 || pos < 7 {
		level.Debug(svc.Logger).Log("API", "UploadPicture", "info", "Bad Picture Data", "owner", owner,
			"DataLength", strconv.Itoa(len(picData)), "sepeatePos", strconv.Itoa(pos))
		return "", errors.New("Bad picture data")
	}

	imgFormat := picData[11 : pos-7]
	realData := picData[pos+1 : len(picData)]

	baseData, err := base64.StdEncoding.DecodeString(realData)
	if err != nil {
		return "", err
	}

	outfileName := picID + "." + imgFormat
	// Local FrontEnd Dev version
	if svc.IsLocalDev {
		outfilePath := path.Join("./data", picFolder, outfileName)

		outputFile, _ := os.Create(outfilePath)
		defer outputFile.Close()

		outputFile.Write(baseData)

		newImageURL := "http://localhost:8000/" + picFolder + "/" + outfileName
		level.Debug(svc.Logger).Log("Picture URL", newImageURL)
		return newImageURL, nil
	}

	storageClient := &http.Client{}
	storageURL := svc.SaveURL + "?userid=" + owner + "&imageid=" + outfileName
	bodyReader := bytes.NewReader(baseData)
	storageReq, err := http.NewRequest("POST", storageURL, bodyReader)
	if err != nil {
		level.Debug(svc.Logger).Log("Storage", storageURL, "err", "request construct fails")
		return "", err
	}

	res, err := storageClient.Do(storageReq)
	if err != nil {
		level.Debug(svc.Logger).Log("Storage", "request fails")
		return "", err
	}

	if res.StatusCode != http.StatusOK {
		return "", errors.New("Upload fails")
	}

	imgReader := bytes.NewReader(baseData)
	// The default image type after image.Decode is jpeg
	img, _, err := image.Decode(imgReader)
	if err != nil {
		return "", err
	}

	_, err = svc.waterMarkAndCache(img, "jpeg", owner+outfileName)
	if err != nil {
		return "", err
	}

	level.Debug(svc.Logger).Log("API", "uploadPicture", "info", "upload picture successfully")

	// construct the memcached url
	return svc.CacheGetURL + "/" + owner + "/" + outfileName, nil
}

// UploadContentFile upload content file to the cloud storage
func (svc *ProductService) UploadContentFile(productData Product) (Product, error) {
	imageID := NSUtil.UniqueID()
	newImageURL, err := svc.uploadPicture(productData.Owner, productData.URL, imageID, "contents")

	newContent := Product{ID: imageID}
	if err != nil {
		return newContent, err
	}

	newContent.URL = newImageURL
	return newContent, nil
}

func (svc *ProductService) newImageId(imageData string) (string, error) {
	newId := NSUtil.GetMd5String(imageData)
	product, _ := svc.GetProductsByID(newId)
	if product.ID == newId {
		return "", errors.New("The product has been uploaded. Please try others.")
	}

	return newId, nil
}

// UploadStyleFile upload style file to the cloud storage
func (svc *ProductService) UploadStyleFile(productData UploadProduct) (Product, error) {
	imageID, err := svc.newImageId(productData.PicData)
	if err != nil {
		level.Error(svc.Logger).Log("API", "UploadStyleFile", "info", err.Error(), "owner", productData.Owner)
		return Product{}, err
	}

	newImageURL, err := svc.uploadPicture(productData.Owner, productData.PicData, imageID, "styles")
	if err != nil {
		level.Error(svc.Logger).Log("API", "UploadStyleFile", "info", err.Error(), "owner", productData.Owner)
		return Product{}, err
	}

	newProduct := Product{ID: imageID}

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
		picID := NSUtil.UniqueID()
		picURL, err := svc.uploadPicture(productData.Owner, pic, picID, "styles")
		if err == nil {
			newProduct.Story.Pictures[index] = picURL
		} else {
			newProduct.Story.Pictures[index] = ""
		}
	}

	// add it to product data to the database
	err = svc.addProduct(newProduct)

	level.Debug(svc.Logger).Log("API", "UploadStyleFile", "info", "Style Upload successful", "owner", productData.Owner)
	return newProduct, nil
}

// UploadStyleFiles upload style file
func (svc *ProductService) UploadStyleFiles(products BatchProducts) (string, error) {
	uploadSize := 0
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
			level.Debug(svc.Logger).Log("Uploads", "fail to upload",
				"info", "Number"+strconv.Itoa(index)+" is failed to upload")
			continue
		}

		uploadSize++
	}

	if uploadSize == 0 {
		level.Error(svc.Logger).Log("API", "UploadStyleFiles", "info", "No file uploaded in batch")
		return "all fails", errors.New("All upload fails")
	}

	level.Debug(svc.Logger).Log("API", "UploadStyleFiles", "info", "Batch file upload successfully",
		"SucessSize", uploadSize, "FailedSize", len(products.PicDatas)-uploadSize)

	return "succeed", nil
}

func (svc *ProductService) addProduct(product Product) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")

	err := c.Insert(product)
	if err != nil {
		if mgo.IsDup(err) {
			level.Error(svc.Logger).Log("API", "addProduct", "info", "Insert Product fails because of duplicated data", "error", err.Error())
			return errors.New("Book with this ISBN already exists")
		}
		level.Error(svc.Logger).Log("API", "addProduct", "info", "Insert product to database fails", "error", err.Error())
		return errors.New("Failed to add a new products")
	}

	return nil
}

// DeleteProduct delele the product by id from the database and cloud storage
func (svc *ProductService) DeleteProduct(productID string) error {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")
	err := c.Remove(bson.M{"id": productID})
	if err != nil {
		level.Error(svc.Logger).Log("API", "DeleteProduct", "err", err.Error())
		// Todo: How to return useful information for the user
		return errors.New("Failed to delete product")
	}

	// Todo: Delete the corresponding data from the cloud storage
	// Need owner and picture id

	level.Debug(svc.Logger).Log("API", "DeleteProduct", "info", "Delete product successfully", "productID", productID)
	return nil
}

// UpdateProduct update the product information by id
func (svc *ProductService) UpdateProduct(productID string, productData UploadProduct) error {
	// Todo: Check the necessary update data
	updateProduct := Product{ID: productID}
	updateProduct.Owner = productData.Owner
	updateProduct.Maker = productData.Maker
	updateProduct.URL = productData.PicData
	updateProduct.Price = productData.Price
	updateProduct.Tags = productData.Tags
	updateProduct.StyleImgURL = productData.StyleImgURL
	updateProduct.Story.Description = productData.Story.Description
	updateProduct.Type = productData.Type
	updateProduct.ChainId = productData.ChainId

	updateProduct.Story.Pictures = productData.Story.Pictures
	for index, pic := range productData.Story.Pictures {
		pos := strings.Index(pic, "http")
		if pos == 0 {
			updateProduct.Story.Pictures[index] = pic
		} else {
			picID := NSUtil.UniqueID()
			picURL, err := svc.uploadPicture(productData.Owner, pic, picID, "styles")
			if err == nil {
				updateProduct.Story.Pictures[index] = picURL
			} else {
				updateProduct.Story.Pictures[index] = ""
			}
		}
	}

	updateData, err := bson.Marshal(&updateProduct)
	if err != nil {
		// Todo: How to add the product information
		level.Error(svc.Logger).Log("API", "UpdateProduct", "info", "bson Marshal Failes", "error", err.Error())
		return errors.New("Failed to update product")
	}
	mData := bson.M{}
	err = bson.Unmarshal(updateData, mData)
	if err != nil {
		// Todo: How to add the product information
		level.Error(svc.Logger).Log("API", "UpdateProduct", "info", "bson UnMarshal fails", "error", err.Error())
		return errors.New("Failed to update product")
	}

	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")
	err = c.Update(bson.M{"id": productID}, bson.M{"$set": mData})
	if err != nil {
		level.Error(svc.Logger).Log("API", "UpdateProduct", "info", "MongoDB update fails", "error", err.Error())
		return errors.New("Failed to update product")
	}

	return nil
}

func (svc *ProductService) UpdateProductAfterTransaction(productId string, newOwner string, newPrice string) error {
	session := svc.Session.Copy()
	defer session.Close()

	updateData := bson.M{"owner": newOwner, "price.value": newPrice}
	c := session.DB("store").C("products")
	err := c.Update(bson.M{"id": productId}, bson.M{"$set": updateData})
	if err != nil {
		level.Error(svc.Logger).Log("API", "Date.Update", "Error", err)
		return errors.New("Failed to update product owner")
	}

	return nil
}

func (svc *ProductService) getQueryBSon(keyvals ...interface{}) (bson.M, error) {
	query := bson.M{}
	querySize := len(keyvals)

	for i := 0; i < querySize-1; i += 2 {
		key := keyvals[i]
		val := keyvals[i+1]
		keyStr, ok := key.(string)
		if !ok || len(keyStr) == 0 {
			level.Debug(svc.Logger).Log("API", "getQueryBSon", "info", "Bad Key as string", "key", key)
			continue
		}

		// Todo: How to check the value and prevent NoSQL injection
		valInfo := reflect.ValueOf(val)
		switch valInfo.Kind() {
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64,
			reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64,
			reflect.Float32, reflect.Float64,
			reflect.String:
			query[keyStr] = val
		case reflect.Array, reflect.Slice:
			if valInfo.Len() == 1 {
				query[keyStr] = valInfo.Index(0).Interface()
			} else {
				//Todo: How to aggregrate array values, $group?
			}
		default:
			level.Debug(svc.Logger).Log("API", "getQueryBSon", "info", "unsupported value type", "type", valInfo.Kind())
		}
	}

	if len(query) == 0 {
		// Todo: how to print the query keyvals
		level.Error(svc.Logger).Log("API", "getQueryBSon", "error", "Bad Query arguments", keyvals)
		return nil, errors.New("Bad Query arguments")
	}

	return query, nil
}

// GetProducts find all the generated products(images)
func (svc *ProductService) GetProducts() ([]Product, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")

	var products []Product
	err := c.Find(bson.M{}).All(&products)
	if err != nil {
		level.Debug(svc.Logger).Log("API", "GetProducts", "info", err.Error())
		return products, errors.New("Database error")
	}

	return products, nil
}

// GetProductsByUser get all the products owner by user
func (svc *ProductService) GetProductsByUser(userID string) ([]Product, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")

	queryParams, err := svc.getQueryBSon("owner", userID)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetProductsByUser", "UserID", userID, "error", err.Error())
		return nil, err
	}

	var products []Product
	err = c.Find(queryParams).All(&products)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetProductsByUser", "UserID", userID, "error", err.Error())
		return nil, errors.New("Database error")
	}

	return products, nil
}

// GetProductsByTags get all the products related to the tags
func (svc *ProductService) GetProductsByTags(tags []string) ([]Product, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("products")

	queryParams, err := svc.getQueryBSon("tags", tags)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetProductsByTags", "tags", tags, "error", err.Error())
		return nil, errors.New("Bad query params")
	}

	var products []Product
	err = c.Find(queryParams).All(&products)
	if err != nil {
		level.Error(svc.Logger).Log("API", "GetProductsByTags", "error", err.Error())
		return nil, errors.New("Database error")
	}
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
		level.Debug(svc.Logger).Log("API", "GetProductsByID", "info", err.Error(), "id", id)
		return Product{}, errors.New("Database error")
	}

	if product.ID != id {
		return Product{}, errors.New("Failed to find product for the id: " + id)
	}

	return product, nil
}

// GetArtists return all the available artists
func (svc *ProductService) GetArtists() ([]Artist, error) {
	session := svc.Session.Copy()
	defer session.Close()

	c := session.DB("store").C("artists")

	var artists []Artist
	err := c.Find(bson.M{}).All(&artists)
	if err != nil {
		level.Debug(svc.Logger).Log("API", "GetArtists", "info", err.Error())
		return artists, errors.New("Database error")
	}

	level.Debug(svc.Logger).Log("API", "GetArtists", "info", "get all artists sucessfully")
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
		level.Debug(svc.Logger).Log("API", "GetHotestArtists", "info", err.Error())
		return artists, errors.New("Database error")
	}

	level.Debug(svc.Logger).Log("API", "GetHotestArtists", "info", "get all the hotest successfully")
	return artists, nil
}

// AddImage add an image file to the memcached
func (svc *ProductService) AddImage(key string, img []byte) error {
	imgItem := memcache.Item{Key: key, Value: img}
	err := svc.CacheClient.Add(&imgItem)
	if err != nil {
		level.Error(svc.Logger).Log("API", "AddImage", "info", "add cached image", "error", err.Error())
	}

	level.Debug(svc.Logger).Log("API", "AddImage", "info", "add cached image successfully")
	return err
}

// GetImage get an image file from the memcached
func (svc *ProductService) GetImage(userID, imageID string) ([]byte, string, error) {
	key := userID + imageID
	mimeType := mime.TypeByExtension("." + "jpg")

	//get key's value
	it, err := svc.CacheClient.Get(key)

	// re add the image again
	if err == memcache.ErrCacheMiss {
		startTime := time.Now()
		storageClient := &http.Client{}
		storageURL := svc.FindURL + "?userid=" + userID + "&imageid=" + imageID

		storageReq, err := http.NewRequest("GET", storageURL, nil)
		if err != nil {
			level.Error(svc.Logger).Log("API", "GetCloudImage", "error", err.Error(), "url", storageURL)
			return nil, "", err
		}
		res, err := storageClient.Do(storageReq)
		if err != nil {
			return nil, "", err
		}

		var urlData map[string]string
		err = json.NewDecoder(res.Body).Decode(&urlData)
		if err != nil {
			level.Error(svc.Logger).Log("API", "GetCloudImageURL", "error", err.Error())
			return nil, "", err
		}

		// get the image data
		imgResponse, err := http.Get(urlData["url"])
		if err != nil {
			level.Error(svc.Logger).Log("API", "GetCloudImageData", "error", err.Error())
			return nil, "", err
		}

		// watermark and cached data
		img, _, err := image.Decode(imgResponse.Body)
		if err != nil {
			level.Error(svc.Logger).Log("API", "ParseImageData", "error", err.Error())
			return nil, "", err
		}

		level.Debug(svc.Logger).Log("API", "GetCloudImage", "info", "Get Cloud Image", "timeDelay", time.Since(startTime))
		startTime = time.Now()
		imgData, err := svc.waterMarkAndCache(img, "jpeg", userID+imageID)

		level.Debug(svc.Logger).Log("API", "WaterMarkAndCache", "info", "Get watermark image and cache", "timeDelay", time.Since(startTime))

		return imgData, mimeType, nil
	}

	if err != nil {
		return nil, "", err
	}

	if string(it.Key) != key {
		return nil, "", errors.New("Unknown Error in memcached for " + key)
	}

	// All the memory cached image item is jpeg
	return it.Value, mimeType, nil
}

func (svc *ProductService) waterMarkAndCache(img image.Image, format, key string) ([]byte, error) {
	watermarkSVC := WaterMark.Service{
		SourceImg: img,
		Text:      "El-force",
		TextColor: color.RGBA{255, 255, 255, 255},
		Scale:     1.0,
		Format:    format,
	}

	markedBytes := make([]byte, 0)
	outputBuffers := bytes.NewBuffer(markedBytes)

	_, err := watermarkSVC.CreateWaterMark(outputBuffers)
	if err != nil {
		level.Debug(svc.Logger).Log("API", "CreateWaterMark", "info", err.Error())
		return nil, err
	}

	// add the memecached item
	err = svc.AddImage(key, outputBuffers.Bytes())
	if err != nil {
		level.Error(svc.Logger).Log("API", "CacheImage", "info", err.Error())
		return nil, err
	}

	return outputBuffers.Bytes(), nil
}

// Search find all the available products by following the key and values
func (svc *ProductService) Search(keyvals map[string]interface{}) ([]Product, error) {
	session := svc.Session.Copy()
	defer session.Close()

	var queryInfo []interface{}
	for key, val := range keyvals {
		queryInfo = append(queryInfo, key)
		queryInfo = append(queryInfo, val)
	}
	c := session.DB("store").C("products")

	queryParams, err := svc.getQueryBSon(queryInfo...)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Search", "error", err.Error())
		return nil, err
	}

	var prods []Product
	err = c.Find(queryParams).All(&prods)
	if err != nil {
		level.Error(svc.Logger).Log("API", "Search", "info", "Find DB fails", "error", err.Error())
	}

	return prods, nil
}
