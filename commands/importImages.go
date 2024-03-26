package cmd

import (
	"crypto/tls"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
	"github.com/wallclockbuilder/stringutil"
)

type FailedImageUpload struct {
	SKU      string
	Handle   string
	ImageURL string
	Reason   string
}

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

var failedImageUploads []FailedImageUpload

func importImages(FilePath string) {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Read provided CSV file and store product info.
	fmt.Println("\nReading products from CSV file...")
	productsFromCSV, err := ReadImageCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("error reading CSV file: %s", err)
		messenger.ExitWithError(err)
	}

	fmt.Println("\nRetrieving products from Vend...")
	productsFromVend := fetchDataForImportImages()

	fmt.Println("\nMatching products from Vend...")
	matchedProducts, err := matchVendProduct(productsFromVend, productsFromCSV)
	if err != nil {
		err = fmt.Errorf("error matching products: %s", err)
		messenger.ExitWithError(err)
	}

	// For each product match, first grab the image from the URL, then post that
	// image to the product on Vend.
	fmt.Println("\nGrabbing images and posting to Vend...")
	uploadedCount := grabAndUploadImage(matchedProducts)

	if len(failedImageUploads) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_image_upload_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedImageUploads)
		if err != nil {
			fmt.Println(color.RedString("\nFailed to write failures to CSV. Printing failures to console instead."))
			for _, failure := range failedImageUploads {
				fmt.Printf("Failed to upload image for SKU: %s, Handle: %s, ImageURL: %s\nReason: %s\n", failure.SKU, failure.Handle, failure.ImageURL, failure.Reason)
			}
		}
	}

	fmt.Printf(color.GreenString("\nFinished! Uploaded %v out of %v products\n"), uploadedCount, len(matchedProducts))

}

func grabAndUploadImage(products []vend.ProductUpload) int {

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(products), "uploading images")
	if err != nil {
		fmt.Printf("error creating progress bar:%s\n", err)
	}

	uploaded := 0
	for _, product := range products {
		bar.Increment()

		imagePath, err := Grab(product)
		if err != nil {
			err = fmt.Errorf("failed to grab image: %s", err)
			failedImageUploads = append(failedImageUploads, FailedImageUpload{
				SKU:      product.SKU,
				Handle:   product.Handle,
				ImageURL: product.ImageURL,
				Reason:   err.Error(),
			})
			continue
		}

		resp, err := vendClient.ImageUploadRequest(product.ID, imagePath)
		if err != nil {
			err = fmt.Errorf("error: %w response: %s", err, string(resp))
			failedImageUploads = append(failedImageUploads, FailedImageUpload{
				SKU:      product.SKU,
				Handle:   product.Handle,
				ImageURL: product.ImageURL,
				Reason:   err.Error(),
			})
		} else {
			uploaded++
		}
	}
	p.Wait()
	return uploaded
}

func fetchDataForImportImages() map[string]vend.Product {
	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("products")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	vc := *vendClient
	_, productsMap, err := vc.Products()

	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("failed while retrieving images: %v", err)
		messenger.ExitWithError(err)
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

	return productsMap
}

func matchVendProduct(productsFromVend map[string]vend.Product, productsFromCSV []vend.ProductUpload) ([]vend.ProductUpload, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(productsFromCSV), "Matching Products")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}
	var products []vend.ProductUpload

	// Loop through each product from the store, and add the ID field
	// to any product from the CSV file that matches.
Match:
	for _, csvProduct := range productsFromCSV {
		bar.Increment()
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
		failedImageUploads = append(failedImageUploads, FailedImageUpload{
			SKU:      csvProduct.SKU,
			Handle:   csvProduct.Handle,
			ImageURL: csvProduct.ImageURL,
			Reason:   "No handle/sku match",
		})
	}

	// Check how many matches we got.
	if len(products) == 0 {
		bar.AbortBar()
		return nil, fmt.Errorf("no product matches - check your handle/sku values")
	}

	p.Wait()
	return products, nil
}

// ReadImageCSV reads the provided CSV file and stores the input as product objects.
func ReadImageCSV(productFilePath string) ([]vend.ProductUpload, error) {
	// SKU and Handle combo should be a unique identifier.
	header := []string{"sku", "handle", "image_url"}

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		err = fmt.Errorf("error creating progress bar:%s", err)
		return nil, err
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Open our provided CSV file.
	file, err := os.Open(productFilePath)
	if err != nil {
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		bar.AbortBar()
		p.Wait()
		return []vend.ProductUpload{}, err
	}
	// Make sure to close at end.
	defer file.Close()

	// Create CSV reader on our file.
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, err := reader.Read()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return []vend.ProductUpload{}, err
	}

	if len(headerRow) > 3 {
		err = fmt.Errorf("header row longer than expected")
		bar.AbortBar()
		p.Wait()
		return []vend.ProductUpload{}, err
	}

	// Check each string in the header row is same as our template.
	for i, row := range headerRow {
		if stringutil.Strip(strings.ToLower(row)) != header[i] {
			bar.AbortBar()
			p.Wait()
			err = fmt.Errorf("mismatched CSV headers, expecting {sku, handle, image_url}")
			return []vend.ProductUpload{}, err
		}
	}

	// Read the rest of the data from the CSV.
	rawData, err := reader.ReadAll()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return []vend.ProductUpload{}, err
	}

	var product vend.ProductUpload
	var productList []vend.ProductUpload
	var rowNumber int

	// Loop through rows and assign them to product.
	for idx, row := range rawData {
		rowNumber++
		product, err = readRow(row)
		if err != nil {
			bar.AbortBar()
			err = fmt.Errorf("error reading row %v from CSV: %s", idx+1, err) // plus one since csvs are 1 indexed
			return productList, err
		}

		// Append each product to our list.
		productList = append(productList, product)
	}

	// Check how many rows we successfully read and stored.
	if len(productList) > 0 {
	} else {
		bar.AbortBar()
		err = fmt.Errorf("no valid products found")
		return productList, err
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

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
			err := errors.New("missing field")
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
		fileName = products.ID
	}

	// Write product data to file
	err = os.WriteFile(fileName, image, 0666)
	if err != nil {
		err = fmt.Errorf("something went wrong writing image to file: %v", err)
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

	// Doing the request.
	res, err := client.Get(url)
	if err != nil {
		err = fmt.Errorf("error fetching image from url")
		return nil, err
	}
	// Make sure response body is closed at end.
	defer res.Body.Close()

	// Check HTTP response.
	err = vend.ResponseCheck(res.StatusCode)
	if err != nil {
		err = fmt.Errorf("%s Status Code: %v", err, res.StatusCode)
		return nil, err
	}

	// Read what we got back.
	body, err := io.ReadAll(res.Body)
	if err != nil {
		err = fmt.Errorf("error while reading response body: %v", err)
		return nil, err
	}

	return body, err
}
