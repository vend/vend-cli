package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
	"github.com/wallclockbuilder/stringutil"
)

// Command config
var importImagesCmd = &cobra.Command{
	Use:   "import-images",
	Short: "Import Product Images",
	Long: fmt.Sprintf(`
This tool requires the Import Images CSV template, you can download it here: http://bit.ly/vendclitemplates

Example:
%s`, color.GreenString("vendcli import-images -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),

	Run: func(cmd *cobra.Command, args []string) {
		importImages(FilePath)
	},
}

func init() {
	// Flags
	importImagesCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	importImagesCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(importImagesCmd)
}

func importImages(FilePath string) {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Read provided CSV file and store product info.
	fmt.Println("\nReading products from CSV file...")
	productsFromCSV, err := ReadImageCSV(FilePath)
	if err != nil {
		log.Fatalf("Error reading CSV file")

	}

	// Get all products from Vend.
	_, productsFromVend, err := vc.Products()
	if err != nil {
		fmt.Printf("Failed to get products")
	}

	// Match products from Vend with those from the provided CSV file.
	matchedProducts := matchVendProduct(productsFromVend, productsFromCSV)
	if err != nil {
		fmt.Printf("Error matching product from Vend to CSV input")
	}

	// For each product match, first grab the image from the URL, then post that
	// image to the product on Vend.
	fmt.Printf("Grabbing images to post to Vend...\n\n")
	for _, product := range matchedProducts {
		imagePath, err := Grab(product)
		if err != nil {
			fmt.Println("Failed to post images to Vend")
			continue
		}
		UploadImage(imagePath, product)
	}

	fmt.Println(color.GreenString("\nFinished!\n"))
}

func matchVendProduct(productsFromVend map[string]vend.Product, productsFromCSV []vend.ProductUpload) []vend.ProductUpload {

	var products []vend.ProductUpload

	// Loop through each product from the store, and add the ID field
	// to any product from the CSV file that matches.
Match:
	for _, csvProduct := range productsFromCSV {
		for _, vendProduct := range productsFromVend {
			// Ignore if any required values are empty.
			if vendProduct.SKU == nil || vendProduct.Handle == nil ||
				csvProduct.SKU == "" || csvProduct.Handle == "" {
				continue
			}
			// Ignore if product deleted.
			if vendProduct.DeletedAt != nil {
				continue
			}
			// Make sure we have a unique handle/sku match, then add product to list.
			if *vendProduct.SKU == csvProduct.SKU &&
				*vendProduct.Handle == csvProduct.Handle {
				products = append(products,
					vend.ProductUpload{
						ID:       *vendProduct.ID,
						Handle:   csvProduct.Handle,
						SKU:      csvProduct.SKU,
						ImageURL: csvProduct.ImageURL,
					})
				continue Match
			}
		}
		// Record product from CSV as error if no match to Vend products.
		err := errors.New("No handle/sku match")
		log.WithError(err).WithFields(log.Fields{
			"type":                  "match",
			"csv_product_sku":       csvProduct.SKU,
			"csv_product_handle":    csvProduct.Handle,
			"csv_product_image_url": csvProduct.ImageURL,
		})
	}

	// Check how many matches we got.
	if len(products) > 0 {
		fmt.Printf(color.GreenString("Matched %v Product out of %v\n", len(products), len(productsFromCSV)))
	} else {
		fmt.Println("No product matches")
		return nil
	}
	return products
}

// ReadImageCSV reads the provided CSV file and stores the input as product objects.
func ReadImageCSV(productFilePath string) ([]vend.ProductUpload, error) {
	// SKU and Handle combo should be a unique identifier.
	header := []string{"sku", "handle", "image_url"}

	// Open our provided CSV file.
	file, err := os.Open(productFilePath)
	if err != nil {
		fmt.Printf("Could not read from CSV file")
		return []vend.ProductUpload{}, err
	}
	// Make sure to close at end.
	defer file.Close()

	// Create CSV reader on our file.
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, err := reader.Read()
	if err != nil {
		fmt.Printf("Failed to read headerow.")
		return []vend.ProductUpload{}, err
	}

	if len(headerRow) > 3 {
		fmt.Printf("Header row longer than expected")
	}

	// Check each string in the header row is same as our template.
	for i, row := range headerRow {
		if stringutil.Strip(strings.ToLower(row)) != header[i] {
			fmt.Println("Mismatched CSV headers, expecting {sku, handle, image_url}")
			return []vend.ProductUpload{}, fmt.Errorf("Mistmatched Headers %v", err)
		}
	}

	// Read the rest of the data from the CSV.
	rawData, err := reader.ReadAll()
	if err != nil {
		return []vend.ProductUpload{}, err
	}

	var product vend.ProductUpload
	var productList []vend.ProductUpload
	var rowNumber int

	// Loop through rows and assign them to product.
	for _, row := range rawData {
		rowNumber++
		product, err = readRow(row)
		if err != nil {
			fmt.Println("Error reading row from CSV")
			continue
		}

		// Append each product to our list.
		productList = append(productList, product)
	}

	// Check how many rows we successfully read and stored.
	if len(productList) > 0 {
	} else {
		fmt.Println("No valid products found")
	}

	return productList, err
}

// Read a single row of a CSV file and check for errors.
func readRow(row []string) (vend.ProductUpload, error) {
	var product vend.ProductUpload

	product.SKU = row[0]
	product.Handle = row[1]
	product.ImageURL = row[2]

	for i := range row {
		if len(row[i]) < 1 {
			err := errors.New("Missing field")
			return product, err
		}
	}
	return product, nil
}

// Grab downloads a product image and writes it to a file.
func Grab(products vend.ProductUpload) (string, error) {

	// Grab the image based on provided URL.
	image, err := urlGet(products.ImageURL)
	if err != nil {
		return "", err
	}

	// Split the URL up to make it easier to grab the file extension.
	parts := strings.Split(products.ImageURL, ".")
	extension := parts[len(parts)-1]
	// If the extension looks about the right length then use it for the
	// filename, otherwise do not.
	var fileName string
	if len(extension) == 3 {
		fileName = fmt.Sprintf("%s.%s", products.ID, extension)
	} else {
		fileName = fmt.Sprintf("%s", products.ID)
	}

	// Write product data to file
	err = ioutil.WriteFile(fileName, image, 0666)
	if err != nil {
		fmt.Printf("Something went wrong writing image to file: %v", err)
		return "", err
	}

	return fileName, err
}

// Get body takes response and returns body.
func urlGet(url string) ([]byte, error) {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	fmt.Printf("Image URL: %v\n", url)

	// Doing the request.
	res, err := client.Get(url)
	if err != nil {
		log.Fatalf("Error performing request")
		return nil, err
	}
	// Make sure response body is closed at end.
	defer res.Body.Close()

	// Check HTTP response.
	if !vend.ResponseCheck(res.StatusCode) {
		fmt.Printf("Status Code: %v", res.StatusCode)
		return nil, err
	}

	// Read what we got back.
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error while reading response body")
		return nil, err
	}

	return body, err
}

// UploadImage uploads a single product image to Vend.
func UploadImage(imagePath string, product vend.ProductUpload) error {
	var err error

	// This checks we actually have an image to post.
	if len(product.ImageURL) > 0 {

		// First grab and save the image from the URL.
		imageURL := fmt.Sprintf("%s", product.ImageURL)

		var body bytes.Buffer
		// Start multipart writer.
		writer := multipart.NewWriter(&body)

		// Key "image" value is the image binary.
		var part io.Writer
		part, err = writer.CreateFormFile("image", imageURL)
		if err != nil {
			fmt.Printf("Error creating multipart form file")
			return err
		}

		// Open image file.
		var file *os.File
		file, err = os.Open(imagePath)
		if err != nil {
			fmt.Printf("Error opening image file")
			return err
		}

		// Make sure file is closed and then removed at end.
		defer file.Close()
		defer os.Remove(imageURL)

		// Copying image binary to form file.
		_, err = io.Copy(part, file)
		if err != nil {
			log.Fatalf("Error copying file for requst body: %s", err)
			return err
		}

		err = writer.Close()
		if err != nil {
			fmt.Printf("Error closing writer")
			return err
		}

		// Create the Vend URL to send our image to.
		url := vendClient.ImageUploadURLFactory(product.ID)

		fmt.Printf("Uploading image to %v, ", product.ID)

		req, _ := http.NewRequest("POST", url, &body)

		// Headers
		req.Header.Set("User-agent", "vend-image-upload")
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", vendClient.Token))

		client := &http.Client{}

		// Make the request.
		var attempt int
		var res *http.Response
		for {
			time.Sleep(time.Second)
			res, err = client.Do(req)
			if err != nil {
				fmt.Printf("Couldnt source image: %s", res.Status)
				// Delays between attempts will be exponentially longer each time.
				attempt++
				delay := vend.BackoffDuration(attempt)
				time.Sleep(delay)
			} else {
				// Ensure that image file is removed after it's uploaded.
				os.Remove(imagePath)
				break
			}
		}

		// Make sure response body is closed at end.
		defer res.Body.Close()

		var resBody []byte
		resBody, err = ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("Error reading response body")
			return err
		}

		// Unmarshal JSON response into our respone struct.
		// from this we can find info about the image status.
		response := vend.ImageUpload{}
		err = json.Unmarshal(resBody, &response)
		if err != nil {
			fmt.Println("error sourcing image - please check the image URL. Image links must be a direct link to the image.")
			os.Exit(1)
			return err
		}

		payload := response.Data

		fmt.Printf(color.GreenString("image created at Position: %v\n\n", *payload.Position))

	}
	return err
}
