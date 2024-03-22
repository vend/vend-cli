package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// Command config
var (
	includeDetails string

	exportImagesCmd = &cobra.Command{
		Use:   "export-images",
		Short: "Export Product Images",
		Long: fmt.Sprintf(`
export-images will export all product images to a CSV file.

Use the "include-details" flag to include extra details in the export. This will include the following fields:
* image position
* image status

WARNING: including extra details issues an API call for each image, so for retailers with a large number of images this will put them at risk of being rated limited. Use with caution.

Example:
%s`, color.GreenString("vendcli export-images -d DOMAINPREFIX -t TOKEN -D INCLUDE-DETAILS")),

		Run: func(cmd *cobra.Command, args []string) {
			getAllImages()
		},
	}
)

func init() {
	// Flags
	exportImagesCmd.Flags().StringVarP(&includeDetails, "include-details", "D", "", "include extra details: true or false")
	exportImagesCmd.MarkFlagRequired("include-details")
	rootCmd.AddCommand(exportImagesCmd)
}

// Run executes the process of grabbing images then writing them to CSV.
func getAllImages() {

	detailsBool := validateDetailsFlag(includeDetails)

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get Images.
	fmt.Println("\nRetrieving Images from Vend...")
	images := fetchDataForImageExport()

	// Write to CSV
	fmt.Println("\nWriting images to CSV file...")
	err := iWriteFile(images, detailsBool)
	if err != nil {
		err = fmt.Errorf("Failed while writing images to CSV: %v", err)
		messenger.ExitWithError(err)
	}

	fmt.Println(color.GreenString("\nFinished!\n"))
}

func fetchDataForImageExport() []vend.Product {
	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("images")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	vc := *vendClient
	products, _, err := vc.Products()

	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("Failed while retrieving images: %v", err)
		messenger.ExitWithError(err)
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

	return products
}

// WriteFile writes image URLs info to file.
func iWriteFile(products []vend.Product, details bool) error {

	vc := *vendClient

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(products), "Writing CSV")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_image_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		err = fmt.Errorf("Failed to create CSV: %v", err)
		messenger.ExitWithError(err)
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	// Write the header line.
	var header []string
	header = append(header, "product_id") // 1
	header = append(header, "image_id")   // 2
	header = append(header, "sku")        // 3
	header = append(header, "handle")     // 4
	header = append(header, "image_url")  // 5

	if details {
		header = append(header, "position") // 6
		header = append(header, "status")   // 7
	}

	// Commit the header.
	writer.Write(header)

	// Now loop through each product object and populate the CSV.
	for _, product := range products {
		bar.Increment()

		var images = product.Images

		var sku, handle, productId string

		if product.ID != nil {
			productId = *product.ID
		}

		if product.SKU != nil {
			sku = *product.SKU
		}
		if product.Handle != nil {
			handle = *product.Handle
		}

		// This will ignore no images since the array will be empty
		for _, image := range images {
			var imageID, url, position, status string

			if image.ID != nil {
				imageID = *image.ID

				if details {
					imageDetails, err := vc.ProductImagesDetails(imageID)
					if err != nil {
						log.Printf(color.RedString("Failed while retrieving image details for image id: %v\n%v", imageID, err))
					}

					if imageDetails.Position != nil {
						position = strconv.FormatInt(*imageDetails.Position, 10)
					}
					if imageDetails.Status != nil {
						status = *imageDetails.Status
					}
				}
			}

			if image.URL != nil {
				url = *image.URL
			}

			var record []string
			record = append(record, productId) // 1
			record = append(record, imageID)   // 2
			record = append(record, sku)       // 3
			record = append(record, handle)    // 4
			record = append(record, url)       // 5

			if details {
				record = append(record, position) // 6
				record = append(record, status)   // 7
			}

			writer.Write(record)
		}

	}
	p.Wait()
	writer.Flush()
	return err
}

func validateDetailsFlag(d string) bool {
	detailsBool, err := strconv.ParseBool(d)
	if err != nil {
		return false // default to false
	}
	return detailsBool
}
