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

var catalogStats vend.CatalogStats

func getAllProducts() {

	//Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")

	//Get Products
	fmt.Println("\nRetrieving Products...")
	products, _, err := vc.Products()
	if err != nil {
		log.Fatalf("Failed retrieving products from Vend %v", err)
	}
	catalogStats.TotalInventory = int64(len(products))

	// Get Outlets
	outlets, outletsMap, err := vc.Outlets()
	if err != nil {
		log.Fatalf("Failed retrieving outlets from Vend %v", err)
	}

	// Get Outlet Taxes
	outletTaxes, err := vc.OutletTaxes()
	if err != nil {
		log.Fatalf("Failed retrieving outlet taxes from Vend %v", err)
	}

	// Get Taxes
	_, taxMaps, err := vc.Taxes()
	if err != nil {
		log.Fatalf("Failed retrieving taxes from Vend %v", err)
	}

	// Get Inventory
	inventoryRecords, err := vc.Inventory()
	if err != nil {
		fmt.Println("Error fetching inventory records")
	}

	// Get Tags
	tagsMap, err := vc.Tags()
	if err != nil {
		fmt.Println("Error fetching tags")
	}

	// Build Maps
	outletTaxesMap := buildOutletTaxesMap(outletTaxes, taxMaps, outlets)
	recordsMap := buildRecordsMap(inventoryRecords, outlets)

	// Write Products to CSV
	fmt.Printf("Writing products to CSV file...\n")
	err = productsWriteFile(products, outlets, outletsMap, recordsMap, outletTaxesMap, tagsMap)
	if err != nil {
		log.Fatalf(color.RedString("Failed writing products to CSV: %v", err))
	}

	// Print happy message, and then display catalog stats
	fmt.Println(color.GreenString("Export Finished!  🎉🎉🎉"))
	printStats()

}

// Creates CSV file and then prints product info to it
func productsWriteFile(products []vend.Product, outlets []vend.Outlet,
	outletsMap map[string][]vend.Outlet, recordsMap map[string]map[string]vend.InventoryRecord,
	outletTaxesMap map[string]map[string]string, tagsMap map[string]string) error {

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
	header = append(header, "primary sku")                 // 2
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
	header = append(header, "tags")                        // 15
	header = append(header, "skus list")                   // 16
	header = append(header, "description")                 // 17
	header = append(header, "count of images")             // 18
	header = append(header, "supply price")                // 19
	header = append(header, "general price excluding tax") // 20
	header = append(header, "loyalty amount")              // 21

	// loop through outlets and list inventory information
	for _, outlet := range outlets {
		header = append(header, fmt.Sprintf("outlet tax: %s", *outlet.Name))      // 22
		header = append(header, fmt.Sprintf("inventory level: %s", *outlet.Name)) // 23
		header = append(header, fmt.Sprintf("current amount: %s", *outlet.Name))  // 24
		header = append(header, fmt.Sprintf("average cost: %s", *outlet.Name))    // 25
		header = append(header, fmt.Sprintf("reorder point: %s", *outlet.Name))   // 26
		header = append(header, fmt.Sprintf("reorder amount: %s", *outlet.Name))  // 27
	}

	header = append(header, "weight unit") // 28
	header = append(header, "weight")      // 29
	header = append(header, "size unit")   // 30
	header = append(header, "length")      // 31
	header = append(header, "width")       // 32
	header = append(header, "height")      // 33
	header = append(header, "active")      // 34
	header = append(header, "created at")  // 35
	header = append(header, "updated at")  // 36
	header = append(header, "deleted at")  // 37
	header = append(header, "version")     // 38

	writer.Write(header)

	// loop through products and write to csv
	for _, product := range products {
		var id, handle, sku, name, productClassification, productType, brandName, supplierName, supplierCode, description,
			tagsList, skuList, imageCount, supplierPrice, priceExcludingTax, loyaltyAmount, weightUnit, weight, sizeUnit,
			length, width, height, active, createdAt, updatedAt, deletedAt, version string

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
			addOne(&catalogStats.CountComposite)
		} else if product.HasVariants {
			productClassification = "PARENT VARIANT"
			addOne(&catalogStats.CountParentVariant)
		} else if product.VariantParentID != nil {
			productClassification = "CHILD VARIANT"
			addOne(&catalogStats.CountChildVariant)
		} else {
			productClassification = "STANDARD"
			addOne(&catalogStats.CountStandard)
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

		if product.LoyaltyAmount != nil {
			loyaltyAmount = fmt.Sprintf("%.2f", *product.LoyaltyAmount)
		} else {
			loyaltyAmount = "default"
		}

		// count number of active and inactive products
		active = strconv.FormatBool(product.Active)
		if product.Active {
			addOne(&catalogStats.CountActive)
		} else {
			addOne(&catalogStats.CountInactive)
		}

		if product.CreatedAt != nil {
			createdAt = *product.CreatedAt
		}

		if product.UpdatedAt != nil {
			updatedAt = *product.UpdatedAt
		}

		if product.DeletedAt != nil {
			deletedAt = *product.DeletedAt
		}

		// create a comma seperated list of tags
		for idx, tag := range product.TagIDs {
			if idx == 0 {
				tagsList = tagsMap[*tag]
			} else {
				tagsList = fmt.Sprintf("%s,%s", tagsList, tagsMap[*tag])
			}
		}

		// create a comma seperated list of skus
		for idx, sku := range product.SKUCodes {
			if sku.Code != nil {
				if idx == 0 {
					skuList = *sku.Code
				} else {
					skuList = fmt.Sprintf("%s, %s", skuList, *sku.Code)
				}
			}
		}

		if product.Description != nil {
			description = *product.Description
		}

		imageCount = strconv.Itoa(len(product.Images))

		if product.WeightUnit != nil {
			weightUnit = *product.WeightUnit
		}

		if product.Weight != nil {
			weight = fmt.Sprintf("%.3f", *product.Weight)
		}

		if product.SizeUnit != nil {
			sizeUnit = *product.SizeUnit
		}

		if product.Length != nil {
			length = fmt.Sprintf("%.3f", *product.Length)
		}

		if product.Width != nil {
			width = fmt.Sprintf("%.3f", *product.Width)
		}

		if product.Height != nil {
			height = fmt.Sprintf("%.3f", *product.Height)
		}

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
		record = append(record, tagsList)              // 15
		record = append(record, skuList)               // 16
		record = append(record, description)           // 17
		record = append(record, imageCount)            // 18
		record = append(record, supplierPrice)         // 19
		record = append(record, priceExcludingTax)     // 20
		record = append(record, loyaltyAmount)         // 21

		// loop through outlets and append inventory & tax information
		for _, outlet := range outlets {

			// outlet tax                                    // 22
			// check if a tax entry exists before setting info
			if taxName, ok := outletTaxesMap[*outlet.ID][*product.ID]; ok {
				record = append(record, taxName)
			} else {
				record = append(record, "Default Tax")
			}

			// check if a record exists then add inventory info
			if invRecord, ok := recordsMap[*outlet.ID][*product.ID]; ok {
				// inventory level                           // 23
				if invRecord.InventoryLevel != nil {
					inventoryLevel := strconv.FormatInt(*invRecord.InventoryLevel, 10)
					record = append(record, inventoryLevel)
				} else {
					record = append(record, "")
				}

				// current amount                            // 24
				if invRecord.CurrentAmount != nil {
					currentAmount := strconv.FormatInt(*invRecord.CurrentAmount, 10)
					record = append(record, currentAmount)
				} else {
					record = append(record, "")
				}

				// average cost                              // 25
				if invRecord.AverageCost != nil {
					averageCost := fmt.Sprintf("%.2f", *invRecord.AverageCost)
					record = append(record, averageCost)
				} else {
					record = append(record, "")
				}

				// reorder point                             // 26
				if invRecord.ReorderPoint != nil {
					reorderPoint := strconv.FormatInt(*invRecord.ReorderPoint, 10)
					record = append(record, reorderPoint)
				} else {
					record = append(record, "")
				}

				// reorderamount                             // 27
				if invRecord.ReorderAmount != nil {
					reorderAmount := strconv.FormatInt(*invRecord.ReorderAmount, 10)
					record = append(record, reorderAmount)
				} else {
					record = append(record, "")
				}

				// even if there isn't an inventory record, we still want to put something in the
				// cell, so our data remains aligned with the header
			} else {
				record = append(record, "") // 23
				record = append(record, "") // 24
				record = append(record, "") // 24
				record = append(record, "") // 26
				record = append(record, "") // 27
			}
		}

		record = append(record, weightUnit) // 28
		record = append(record, weight)     // 29
		record = append(record, sizeUnit)   // 30
		record = append(record, length)     // 31
		record = append(record, width)      // 32
		record = append(record, height)     // 33
		record = append(record, active)     // 34
		record = append(record, createdAt)  // 35
		record = append(record, updatedAt)  // 36
		record = append(record, deletedAt)  // 37
		record = append(record, version)    // 38

		writer.Write(record)
	}
	writer.Flush()
	return err
}

// build hash table so inventory records can be accessed quickly
func buildRecordsMap(inventoryRecords []vend.InventoryRecord, outlets []vend.Outlet) map[string]map[string]vend.InventoryRecord {
	var recordsMap = map[string]map[string]vend.InventoryRecord{}

	// set outlet maps, first
	for _, outlet := range outlets {
		recordsMap[*outlet.ID] = map[string]vend.InventoryRecord{}
	}

	// set product maps, second
	for _, record := range inventoryRecords {

		// check that we don't have any nil values
		if record.OutletID != nil && record.ProductID != nil {

			// make sure a give Outlet map exists and set product map
			if _, ok := recordsMap[*record.OutletID]; ok {
				recordsMap[*record.OutletID][*record.ProductID] = record
			}
		}
	}

	return recordsMap
}

// build hash table so tax display name for given product/outlet pair can be quickly accessed
func buildOutletTaxesMap(outletTaxes []vend.OutletTaxes, TaxesMap map[string]string, outlets []vend.Outlet) map[string]map[string]string {
	var outletTaxesMap = map[string]map[string]string{}

	// set outlet maps, first
	for _, outlet := range outlets {
		outletTaxesMap[*outlet.ID] = map[string]string{}
	}

	// set product map, second
	for _, outletTax := range outletTaxes {

		// check that we don't have any nil values
		if outletTax.OutletID != nil && outletTax.ProductID != nil && outletTax.TaxID != nil {

			// make sure a give Outlet map exists, a Tax map exists, then set product map
			if _, ok := outletTaxesMap[*outletTax.OutletID]; ok {
				if taxDisplayName, ok := TaxesMap[*outletTax.TaxID]; ok {
					outletTaxesMap[*outletTax.OutletID][*outletTax.ProductID] = taxDisplayName
				}
			}
		}
	}
	return outletTaxesMap
}

// helper function for stats
func addOne(num *int64) {
	*num = *num + 1
}

func printStats() {
	fmt.Printf(`
Catalog Stats...
	Total Products: %s

Product Classifications:
	Composite Items: %s
	Parent Variants: %s
	Child Variants: %s
	Standard Products: %s
	
Active:
	Active Products: %s
	InActive Products: %s

`,
		color.GreenString(strconv.FormatInt(catalogStats.TotalInventory, 10)),
		color.GreenString(strconv.FormatInt(catalogStats.CountComposite, 10)),
		color.GreenString(strconv.FormatInt(catalogStats.CountParentVariant, 10)),
		color.GreenString(strconv.FormatInt(catalogStats.CountChildVariant, 10)),
		color.GreenString(strconv.FormatInt(catalogStats.CountStandard, 10)),
		color.GreenString(strconv.FormatInt(catalogStats.CountActive, 10)),
		color.GreenString(strconv.FormatInt(catalogStats.CountInactive, 10)))
}
