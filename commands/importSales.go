package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// import sales command
var (
	filePath            string
	overwrite           string
	mode                string
	timeZoneImportSales string

	importSalesCmd = &cobra.Command{
		Use:   "import-sales",
		Short: "Import Sales",
		Long: fmt.Sprintf(`
	Example:
	%s`, color.GreenString("vendcli import-sales -d DOMAINPREFIX -t TOKEN -f FILENAME.csv -m MODE -o OVERWRITE")),
		Run: func(cmd *cobra.Command, args []string) {
			importSales()
		},
	}
)

func init() {
	// Flag
	importSalesCmd.Flags().StringVarP(&filePath, "filename", "f", "", "The name of your file: filename.csv")
	importSalesCmd.MarkFlagRequired("filename")

	importSalesCmd.Flags().StringVarP(&overwrite, "overwrite", "o", "", "overwrite sales: true or false")

	importSalesCmd.Flags().StringVarP(&mode, "mode", "m", "parse", "modes: parse, post")
	importSalesCmd.Flags().StringVarP(&timeZoneImportSales, "Timezone", "z", "", "Timezone of the store in zoneinfo format.")

	rootCmd.AddCommand(importSalesCmd)

}

func importSales() {

	overwriteBool := parseOverwriteFlag(overwrite)
	isPostMode := validateModeFlag(mode)

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// fetch the jsons from file and store them into interfaces
	erroredSales, err := readJSONFile(filePath)
	if err != nil {
		fmt.Printf("Error reading json file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	if isPostMode {
		fmt.Printf("\nRunning command in post mode with overwrite set to %t\n", overwriteBool)
		if overwriteBool {
			postSales(erroredSales)
		} else {
			checkedBeforePosting(erroredSales)
		}
	} else {
		fmt.Printf("\nRunning command in parse mode\n")
		// 1970-01-01T00:00:00Z is just a dummy date to validate the timezone
		if len(timeZoneImportSales) == 0 {
			fmt.Println("Timezone is required for parse mode")
			panic(vend.Exit{1})
		} else {
			validateTimeZone("1970-01-01T00:00:00Z", timeZone)
			parseSales(erroredSales)
		}
	}
}

// read JSONFile fetches the errored sales from file and stores them into an array of vend.Sale9 structs
func readJSONFile(jsonFilePath string) ([]vend.Sale9, error) {
	// read file
	// Open our jsonFile
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		fmt.Printf("Error opening json file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	// defer closing jsonFile so we can parse it
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array
	fileContent, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Error reading json file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	// parse json file
	data, err := parseJsonFile(fileContent)
	if err != nil {
		fmt.Printf("Error parsing json file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	// convert generic interface to []vend.Sale9
	switch convertedData := data.(type) {
	case []vend.Sale9:
		return convertedData, nil
	case vend.RegisterSale9:
		return convertedData.RegisterSale9, nil
	default:
		return nil, fmt.Errorf("unknown JSON format")
	}
}

// parseJsonFile guesses the format of the json and tries to unmarshal into the correct struct
// we're using keys to guess the format
func parseJsonFile(fileContent []byte) (interface{}, error) {

	var jsonData interface{}
	if err := json.Unmarshal(fileContent, &jsonData); err != nil {
		return nil, err
	}

	switch data := jsonData.(type) {

	case []interface{}:
		if len(data) > 0 {
			if m, ok := data[0].(map[string]interface{}); ok {
				// iOS errored sales format
				if _, ok := m["register_sale_products"]; ok {
					var sales []vend.Sale9
					err := json.Unmarshal(fileContent, &sales)
					if err != nil {
						return nil, err
					} else {
						return sales, nil
					}
				} else {
					return nil, fmt.Errorf("unknown JSON format")
				}
			} else {
				return nil, fmt.Errorf("unknown JSON format")
			}
		} else {
			return nil, fmt.Errorf("unknown JSON format")
		}

	case map[string]interface{}:
		// Legacy 0.9 sales format
		if _, ok := data["register_sales"]; ok {
			var saleResponse vend.RegisterSale9
			err := json.Unmarshal(fileContent, &saleResponse)
			if err != nil {
				return nil, err
			} else {
				return saleResponse.RegisterSale9, nil
			}
			// Web errored sales format
		} else if _, ok := data["erroredSales"]; ok {
			return nil, fmt.Errorf("web errored sales format is not supported")
		} else {
			return nil, fmt.Errorf("unknown JSON format")
		}
	default:
		return nil, fmt.Errorf("unknown JSON format")
	}
}

func parseSales(sales9 []vend.Sale9) {
	sales, err := vend.ConvertSale9ToSale(sales9)
	if err != nil {
		fmt.Printf("Error converting sales:\n%s\n", err)
		panic(vend.Exit{1})
	}
	sortBySaleDate(sales)

	// Get other data
	registers, users, customers, customerGroupMap, products := GetVendDataForSalesReport(*vendClient)

	// Create report
	file, err := createErredSalesReport()
	if err != nil {
		fmt.Printf("Error creating CSV file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	defer file.Close()

	file = addSalesReportHeader(file)

	fmt.Println("Writing sales report...")
	file = writeSalesReport(file, registers, users, customers, customerGroupMap, products, sales, DomainPrefix, timeZoneImportSales)

	fmt.Printf("Sales report created: %s\n", file.Name())
}

func postSales(sales []vend.Sale9) {
	fmt.Println("Posting Sales!")

	for idx, sale := range sales {
		if sale.ID != nil {
			err := postSale(sale)
			if err != nil {
				fmt.Printf("Error posting sale %s: %s\n", *sale.ID, err)
				continue
			} else {
				fmt.Printf("Sale %s posted successfully!\n", *sale.ID)
			}
		} else {
			fmt.Printf("Sale ID of sale number: %v in list is nil. Skipping..", idx)
		}
	}
}

func checkedBeforePosting(sales []vend.Sale9) {
	fmt.Println("Checking sales before posting!")

	for idx, sale := range sales {

		if sale.ID != nil {
			saleID := *sale.ID

			exists, err := saleExists(saleID)
			if err != nil {
				fmt.Printf("Error checking if sale %s exists: %s\n", saleID, err)
				fmt.Println("Skipping..")
				continue
			}

			if exists {
				fmt.Printf("Sale %s already exists. Skipping..\n", saleID)
			} else {
				err = postSale(sale)
				if err != nil {
					fmt.Printf("Error posting sale %s: %s\n", saleID, err)
					continue
				} else {
					fmt.Printf("Sale %s posted successfully!\n", saleID)
				}
			}
		} else {
			fmt.Printf("Sale ID of sale number: %v in list is nil. Skipping..", idx)
		}
	}
}

func postSale(sale vend.Sale9) error {
	fmt.Printf("Posting sale %s\n", *sale.ID)

	vc := *vendClient
	url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales", DomainPrefix)

	_, err := vc.MakeRequest("POST", url, sale)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// saleExists checks the endpoint /register_sales/{id} for an error
func saleExists(id string) (bool, error) {

	vc := *vendClient
	saleResponse := vend.RegisterSales{}

	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales/%s", DomainPrefix, id)

	// Make the request
	res, err := vc.MakeRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	// Unmarshal JSON Response
	err = json.Unmarshal(res, &saleResponse)
	if err != nil {
		return false, err
	}

	// in testing, an invalid uuid returns a 200 with an empty array
	if len(saleResponse.RegisterSales) > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func parseOverwriteFlag(o string) bool {
	overwriteBool, err := strconv.ParseBool(o)
	if err != nil {
		// default to false
		overwriteBool = false
	}

	return overwriteBool
}

func createErredSalesReport() (*os.File, error) {

	fileName := fmt.Sprintf("erred_sales%s_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		fmt.Printf("Error creating CSV file: %s", err)
		panic(vend.Exit{1})
	}

	writer := csv.NewWriter(file)

	var warningMessage []string
	warningMessage = append(warningMessage, "NOTE:")
	warningMessage = append(warningMessage, "This file contains sales that received an error when posting.")
	warningMessage = append(warningMessage, "These sales have not been posted")

	writer.Write(warningMessage)
	writer.Flush()

	return file, err
}

func validateModeFlag(m string) bool {
	m = strings.ToLower(m)
	if m == "parse" {
		return false
	} else if m == "post" {
		return true
	} else {
		fmt.Printf("'%s' is not a valid option for -m mode. Mode should be 'parse' or 'post'\n", m)
		panic(vend.Exit{1})
	}
}
