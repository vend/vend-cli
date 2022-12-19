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

	// Get Outlets
	outlets, outletsMap, err := vc.Outlets()
	if err != nil {
		log.Fatalf("Failed retrieving outlets from Vend %v", err)
	}

	// Get Inventory
	inventoryRecords, err := vc.Inventory()
	if err != nil {
		fmt.Println("Error fetching inventory records")
	}

	recordsMap := buildRecordsMap(inventoryRecords, outlets)

	// Write Products to CSV
	fmt.Println("Writing products to CSV file...")
	err = productsWriteFile(products, outlets, outletsMap, recordsMap)
	if err != nil {
		log.Fatalf(color.RedString("Failed writing products to CSV: %v", err))
	}
	fmt.Println(color.GreenString("\nExported %v products  🎉\n", len(products)))
}

func productsWriteFile(products []vend.Product, outlets []vend.Outlet,
	outletsMap map[string][]vend.Outlet, recordsMap map[string]map[string]vend.InventoryRecord) error {

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
	header = append(header, "id")                          // 0
	header = append(header, "handle")                      // 1
	header = append(header, "sku")                         // 2
	header = append(header, "name")                        // 3
	header = append(header, "product classification")      // 4
	header = append(header, "option 1 name")               // 5
	header = append(header, "option 1 value")              // 6
	header = append(header, "option 2 name")               // 7
	header = append(header, "option 2 value")              // 8
	header = append(header, "option 3 name")               // 9
	header = append(header, "option 3 value")              // 10
	header = append(header, "product type")                // 11
	header = append(header, "brand name")                  // 12
	header = append(header, "supplier name")               // 13
	header = append(header, "supplier code")               // 14
	header = append(header, "description")                 // 15
	header = append(header, "count of images")             // 16
	header = append(header, "supply price")                // 17
	header = append(header, "general price excluding tax") // 18

	// loop through outlets and list inventory information
	for _, outlet := range outlets {
		header = append(header, fmt.Sprintf("inventory level: %s", *outlet.Name)) // 19
		header = append(header, fmt.Sprintf("current amount: %s", *outlet.Name))  // 20
		header = append(header, fmt.Sprintf("average cost: %s", *outlet.Name))    // 21
		header = append(header, fmt.Sprintf("reorder point: %s", *outlet.Name))   // 22
		header = append(header, fmt.Sprintf("reorder amount: %s", *outlet.Name))  // 23
	}

	header = append(header, "active")     // 24
	header = append(header, "created at") // 25
	header = append(header, "updated at") // 26
	header = append(header, "deleted at") // 27
	header = append(header, "version")    // 28

	writer.Write(header)

	// loop through products and write to csv
	for _, product := range products {
		var id, handle, sku, name, productClassification, productType, brandName, supplierName, supplierCode, description,
			imageCount, supplierPrice, priceExcludingTax, active, createdAt, updatedAt, deletedAt, version string

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

		if product.IsComposite {
			productClassification = "COMPOSITE"
		} else if product.HasVariants {
			productClassification = "PARENT VARIANT"
		} else if product.VariantParentID != nil {
			productClassification = "CHILD VARIANT"
		} else {
			productClassification = "STANDARD"
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
			if supplier.Price != nil {
				supplierPrice = fmt.Sprintf("%.2f", *supplier.Price)
			}

		}

		if product.PriceExcludingTax != nil {
			priceExcludingTax = fmt.Sprintf("%.2f", *product.PriceExcludingTax)
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

		imageCount = strconv.Itoa(len(product.Images))

		if product.Version != nil {
			version = strconv.FormatInt(*product.Version, 10)
		}

		var record []string
		record = append(record, id)                    // 0
		record = append(record, handle)                // 1
		record = append(record, sku)                   // 2
		record = append(record, name)                  // 3
		record = append(record, productClassification) // 4
		record = append(record, variantName[0])        // 5
		record = append(record, variantValue[0])       // 6
		record = append(record, variantName[1])        // 7
		record = append(record, variantValue[1])       // 8
		record = append(record, variantName[2])        // 9
		record = append(record, variantValue[2])       // 10
		record = append(record, productType)           // 11
		record = append(record, brandName)             // 12
		record = append(record, supplierName)          // 13
		record = append(record, supplierCode)          // 14
		record = append(record, description)           // 15
		record = append(record, imageCount)            // 16
		record = append(record, supplierPrice)         // 17
		record = append(record, priceExcludingTax)     // 18

		// loop through outlets and append inventory information
		for _, outlet := range outlets {
			if invRecord, ok := recordsMap[*outlet.ID][*product.ID]; ok {
				// inventory level                           // 19
				if invRecord.InventoryLevel != nil {
					inventoryLevel := strconv.FormatInt(*invRecord.InventoryLevel, 10)
					record = append(record, inventoryLevel)
				} else {
					record = append(record, "")
				}

				// current amount                            // 20
				if invRecord.CurrentAmount != nil {
					currentAmount := strconv.FormatInt(*invRecord.CurrentAmount, 10)
					record = append(record, currentAmount)
				} else {
					record = append(record, "")
				}

				// average cost                              // 21
				if invRecord.AverageCost != nil {
					averageCost := fmt.Sprintf("%.2f", *invRecord.AverageCost)
					record = append(record, averageCost)
				} else {
					record = append(record, "")
				}

				// reorder point                             // 22
				if invRecord.ReorderPoint != nil {
					reorderPoint := strconv.FormatInt(*invRecord.ReorderPoint, 10)
					record = append(record, reorderPoint)
				} else {
					record = append(record, "")
				}

				// reorderamount                             // 23
				if invRecord.ReorderAmount != nil {
					reorderAmount := strconv.FormatInt(*invRecord.ReorderAmount, 10)
					record = append(record, reorderAmount)
				} else {
					record = append(record, "")
				}

				// even if there isn't an inventory record, we still want to put something in the
				// cell, so our data remains aligned with the header
			} else {
				record = append(record, "") // 19
				record = append(record, "") // 20
				record = append(record, "") // 21
				record = append(record, "") // 22
				record = append(record, "") // 23
			}
		}
		record = append(record, active)    // 24
		record = append(record, createdAt) // 25
		record = append(record, updatedAt) // 26
		record = append(record, deletedAt) // 27
		record = append(record, version)   // 28

		writer.Write(record)
	}
	writer.Flush()
	return err
}

// builds hash table so inventory records can be accessed quickly
func buildRecordsMap(inventoryRecords []vend.InventoryRecord, outlets []vend.Outlet) map[string]map[string]vend.InventoryRecord {
	var recordsMap = map[string]map[string]vend.InventoryRecord{}

	for _, outlet := range outlets {
		recordsMap[*outlet.ID] = map[string]vend.InventoryRecord{}
	}

	for _, record := range inventoryRecords {

		if _, ok := recordsMap[*record.OutletID]; ok {
			recordsMap[*record.OutletID][*record.ProductID] = record
		}
	}

	return recordsMap
}
