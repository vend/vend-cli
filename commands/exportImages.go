package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// Command config
var exportImagesCmd = &cobra.Command{
	Use:   "export-images",
	Short: "Export Product Images",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-images -d DOMAINPREFIX -t TOKEN")),

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
	fmt.Println("\nRetrieving Images from Vend...")
	images, _, err := vc.Products()
	if err != nil {
		log.Printf(color.RedString("Failed while retrieving images: %v", err))
		panic(vend.Exit{1})
	}

	// Write to CSV
	fmt.Println("Writing images to CSV file...")
	err = iWriteFile(images)
	if err != nil {
		log.Printf(color.RedString("Failed while writing images to CSV: %v", err))
		panic(vend.Exit{1})
	}

	fmt.Println(color.GreenString("\nFinished!\n"))
}

// WriteFile writes image URLs info to file.
func iWriteFile(products []vend.Product) error {

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_image_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Printf("Failed to create CSV: %v", err)
		panic(vend.Exit{1})
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

		var images = product.Images

		var sku, handle string

		if product.SKU != nil {
			sku = *product.SKU
		}
		if product.Handle != nil {
			handle = *product.Handle
		}

		// This will ignore no images since the array will be empty
		for _, image := range images {
			var record []string
			record = append(record, sku)
			record = append(record, handle)
			record = append(record, *image.URL)
			writer.Write(record)
		}

	}

	writer.Flush()
	return err
}
