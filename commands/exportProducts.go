package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// Command config
var exportProductsCmd = &cobra.Command{
	Use:   "export-products",
	Short: "Export Products",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-products -d DOMAINPREFIX -t TOKEN ")),

	Run: func(cmd *cobra.Command, args []string) {
		getAllProducts()
	},
}

func init() {
	// Flags
	rootCmd.AddCommand(exportProductsCmd)
}

func getAllProducts() {

	//Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")

	//Get Products
	fmt.Println("\nRetrieving Products...")
	products, _, err := vc.Products()
	if err != nil {
		log.Fatalf("Failed retrieving products from Vend %v", err)
	}

	// Write Products to CSV
	fmt.Println("Writing products to CSV file...")
	err = productsWriteFile(products)
	if err != nil {
		log.Fatalf(color.RedString("Failed writing products to CSV: %v", err))
	}
	fmt.Println(color.GreenString("\nExported %v products  🎉\n", len(products)))
}

func productsWriteFile(products []vend.Product) error {

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_product_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		return err
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	var header []string
	header = append(header, "id")             // 0
	header = append(header, "handle")         // 1
	header = append(header, "sku")            // 2
	header = append(header, "name")           // 3
	header = append(header, "option 1 name")  // 4
	header = append(header, "option 1 value") // 5
	header = append(header, "option 2 name")  // 6
	header = append(header, "option 2 value") // 7
	header = append(header, "option 3 name")  // 8
	header = append(header, "option 3 value") // 9
	header = append(header, "product type")   // 10
	header = append(header, "brand name")     // 11
	header = append(header, "supplier name")  // 12
	header = append(header, "supplier code")  // 13
	header = append(header, "active")         // 14
	header = append(header, "description")    // 15
	header = append(header, "created at")     // 16
	header = append(header, "updated at")     // 17
	header = append(header, "deleted at")     // 18
	header = append(header, "version")        // 19

	writer.Write(header)

	// loop through products and write to csv
	for _, product := range products {
		var id, handle, sku, name, productType, brandName, supplierName, supplierCode, active, createdAt,
			updatedAt, deletedAt, description, version string

		var variantName, variantValue [3]string

		if product.ID != nil {
			id = *product.ID
		}

		if product.Handle != nil {
			handle = *product.Handle
		}

		if product.SKU != nil {
			sku = *product.SKU
		}

		if product.Name != nil {
			name = *product.Name
		}

		// set variant option fields
		for idx, variant := range product.VariantOptions {
			if variant.Name != nil {
				variantName[idx] = *variant.Name
			}
			if variant.Value != nil {
				variantValue[idx] = *variant.Value
			}
		}

		if product.Type.Name != nil {
			productType = *product.Type.Name
		}

		if product.Brand.Name != nil {
			brandName = *product.Brand.Name
		}

		if len(product.ProductSuppliers) > 0 {
			supplier := product.ProductSuppliers[0]
			if supplier.Code != nil {
				supplierCode = *supplier.Code
			}
			if supplier.SupplierName != nil {
				supplierName = *supplier.SupplierName
			}
		}

		active = strconv.FormatBool(product.Active)

		if product.CreatedAt != nil {
			createdAt = *product.CreatedAt
		}

		if product.UpdatedAt != nil {
			updatedAt = *product.UpdatedAt
		}

		if product.DeletedAt != nil {
			deletedAt = *product.DeletedAt
		}

		if product.Description != nil {
			description = *product.Description
		}

		if product.Version != nil {
			version = strconv.FormatInt(*product.Version, 10)
		}

		var record []string
		record = append(record, id)              //0
		record = append(record, handle)          // 1
		record = append(record, sku)             // 2
		record = append(record, name)            // 3
		record = append(record, variantName[0])  // 4
		record = append(record, variantValue[0]) // 5
		record = append(record, variantName[1])  // 6
		record = append(record, variantValue[1]) // 7
		record = append(record, variantName[2])  // 8
		record = append(record, variantValue[2]) // 9
		record = append(record, productType)     // 10
		record = append(record, brandName)       // 11
		record = append(record, supplierName)    // 12
		record = append(record, supplierCode)    // 13
		record = append(record, active)          // 14
		record = append(record, description)     // 15
		record = append(record, createdAt)       // 16
		record = append(record, updatedAt)       // 17
		record = append(record, deletedAt)       // 18
		record = append(record, version)         // 19

		writer.Write(record)
	}
	writer.Flush()
	return err
}
