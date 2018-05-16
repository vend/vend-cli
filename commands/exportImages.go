package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

// exportImagesCmd represents the exportImages command
var exportImagesCmd = &cobra.Command{
	Use:   "export-images",
	Short: "Export Product Images",
	Long: `
Example:
vend export-images -d DOMAINPREFIX -t TOKEN`,

	Run: func(cmd *cobra.Command, args []string) {
		getAllImages()
	},
}

func init() {
	rootCmd.AddCommand(exportImagesCmd)
}

// Run executes the process of grabbing images then writing them to CSV.
func getAllImages() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")

	// Get Images.
	fmt.Println("Retrieving Images from Vend...")
	images, _, err := vc.Products()
	if err != nil {
		log.Fatalf("Failed while retrieving images: %v", err)
	}

	// Write to CSV
	fmt.Println("Writing images to CSV file...")
	err = iWriteFile(images, DomainPrefix)
	if err != nil {
		log.Fatalf("Failed while writing images to CSV: %v", err)
	}

	fmt.Println("Finished!")
}

// WriteFile writes image URLs info to file.
func iWriteFile(products []vend.Product, DomainPrefix string) error {

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_image_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Fatalf("Failed to create CSV: %v", err)
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	// Write the header line.
	var header []string
	header = append(header, "sku")
	header = append(header, "handle")
	header = append(header, "image_url")

	// Commit the header.
	writer.Write(header)

	// Now loop through each product object and populate the CSV.
	for _, product := range products {

		// Ignore Vend placeholder image
		if strings.HasPrefix(*product.ImageURL, "https://secure.vendhq.com/images/placeholder") {
			continue
		}

		var sku, handle, imageURL string

		if product.SKU != nil {
			sku = *product.SKU
		}
		if product.Handle != nil {
			handle = *product.Handle
		}
		if product.ImageURL != nil {
			imageURL = *product.ImageURL
		}

		var record []string
		record = append(record, sku)
		record = append(record, handle)
		record = append(record, imageURL)
		writer.Write(record)
	}

	writer.Flush()
	return err
}
