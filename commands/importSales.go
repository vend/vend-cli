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

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

type FailedSalePostRequest struct {
	SaleID string
	Reason string
}

// import sales command
var (
	filePath               string
	overwrite              string
	mode                   string
	timeZoneImportSales    string
	failedSalePostRequests []FailedSalePostRequest

	importSalesCmd = &cobra.Command{
		Use:   "import-sales",
		Short: "Import Sales",
		Long: fmt.Sprintf(`

import-sales command is used to import sales from a json file to Vend.

The most popular use case is for errored sales, but you can also use this command to edit sales.

It offers two modes, "parse" and "post":
* Parse mode will parse the sales and create a CSV report of the sales. This report matches the report of export-sales command.
* Post mode will post the sales to Vend. 

When posting you can choose whether or not you'd like to "overwrite" existing sales: 
* If overwrite is set to "false" (the default) it will check if the sale already exists before posting. If the sale exists, it will skip it.
* If overwrite is set to "true" it will post the sale regardless of whether it already exists or not.

Using overwrite "true" is useful when you want to change a field of an existing sale that you can not edit through the UI. For example, if you want to change the sales person attribution, customer_id, or tax label of a sale that already exists. To use this command in this way, you would first export the sale from the api/register_sales endpoint using postman, edit the fields in the json, save this to a file, and then import it back in with overwrite set to "true".

When using this command for errored sales, it is highly recommended that you use parse mode first to check the sales before posting.

Example:
	%s`, color.GreenString("vendcli import-sales -d DOMAINPREFIX -t TOKEN -f FILENAME.json -m MODE -o OVERWRITE -z TIMEZONE")),
		Run: func(cmd *cobra.Command, args []string) {
			importSales()
		},
	}
)

func init() {
	// Flag
	importSalesCmd.Flags().StringVarP(&filePath, "filename", "f", "", "The name of your file: filename.json")
	importSalesCmd.MarkFlagRequired("filename")

	importSalesCmd.Flags().StringVarP(&overwrite, "overwrite", "o", "", "overwrite sales: true or false")

	importSalesCmd.Flags().StringVarP(&mode, "mode", "m", "parse", "modes: parse, post")
	importSalesCmd.Flags().StringVarP(&timeZoneImportSales, "Timezone", "z", "", "Timezone of the store in zoneinfo format.")

	rootCmd.AddCommand(importSalesCmd)

}

func importSales() {

	overwriteBool := parseOverwriteFlag(overwrite)
	isPostMode := validateModeFlag(mode)
	if isPostMode {
		fmt.Printf("\nRunning Command in %s mode\n", color.RedString("POST"))
		fmt.Printf("Overwrite: %s\n", prettyOverwriteBool(overwriteBool))
	} else {
		fmt.Printf("\nRunning Command in %s mode\n", color.GreenString("PARSE"))
	}

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// fetch the jsons from file and store them into interfaces
	fmt.Println("\nReading JSON file...")
	erroredSales, err := readJSONFile(filePath)
	if err != nil {
		err = fmt.Errorf("error reading json file: %s ", err)
		messenger.ExitWithError(err)
	}

	if isPostMode {
		if overwriteBool {
			postSales(erroredSales)
		} else {
			checkedBeforePosting(erroredSales)
		}
	} else {
		// 1970-01-01T00:00:00Z is just a dummy date to validate the timezone
		if len(timeZoneImportSales) == 0 {
			err = fmt.Errorf("timezone is required for parse mode")
			messenger.ExitWithError(err)
		} else {
			validateTimeZone("1970-01-01T00:00:00Z", timeZone)
			parseSales(erroredSales)
		}
	}

	if len(failedSalePostRequests) > 0 {
		fmt.Println(color.RedString("\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_post_sale_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedSalePostRequests)
		if err != nil {
			messenger.ExitWithError(err)
			return
		}
	}

	fmt.Println(color.GreenString("\nFinished! 🎉\n"))
}

// read JSONFile fetches the errored sales from file and stores them into an array of vend.Sale9 structs
func readJSONFile(jsonFilePath string) ([]vend.Sale9, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading JSON")
	if err != nil {
		err = fmt.Errorf("error creating progress bar: %s", err)
		return nil, err
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Open our jsonFile
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("error opening json file: %s ", err)
		messenger.ExitWithError(err)
	}

	// defer closing jsonFile so we can parse it
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array
	fileContent, err := io.ReadAll(jsonFile)
	if err != nil {
		bar.AbortBar()
		err = fmt.Errorf("error reading json file: %s ", err)
		messenger.ExitWithError(err)
	}

	// parse json file
	data, err := parseJsonFile(fileContent)
	if err != nil {
		bar.AbortBar()
		err = fmt.Errorf("error parsing json file: %s", err)
		messenger.ExitWithError(err)
	}

	// convert generic interface to []vend.Sale9
	switch convertedData := data.(type) {
	case []vend.Sale9:
		bar.SetIndeterminateBarComplete()
		p.Wait()
		return convertedData, nil
	case vend.RegisterSale9:
		bar.SetIndeterminateBarComplete()
		p.Wait()
		return convertedData.RegisterSale9, nil
	default:
		bar.AbortBar()
		p.Wait()
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
		err = fmt.Errorf("error converting sales: %s ", err)
		messenger.ExitWithError(err)
	}
	sortBySaleDate(sales)

	// Get other data
	registers, users, customers, customerGroupMap, products := GetVendDataForSalesReport(*vendClient)

	// Create report
	file, err := createErredSalesReport()
	if err != nil {
		err = fmt.Errorf("error creating CSV file: %s ", err)
		messenger.ExitWithError(err)
	}

	defer file.Close()

	file = addSalesReportHeader(file)

	fmt.Println("\nWriting sales report...")
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(sales), "Write Report")
	if err != nil {
		fmt.Println(err)
	}
	file = writeSalesReport(file, bar, registers, users, customers, customerGroupMap, products, sales, DomainPrefix, timeZoneImportSales)
	p.Wait()

	fmt.Printf("\nSales report created: %s\n", file.Name())
}

func postSales(sales []vend.Sale9) {
	fmt.Println("\nPosting Sales..")

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(sales), "posting sales")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	for idx, sale := range sales {
		bar.Increment()
		if sale.ID != nil {
			err := postSale(sale)
			if err != nil {
				err = fmt.Errorf("error posting sale: %s", err)
				failedSalePostRequests = append(failedSalePostRequests, FailedSalePostRequest{
					SaleID: *sale.ID,
					Reason: err.Error(),
				})
				continue
			}
		} else {
			err = fmt.Errorf("sale ID of sale number: %v is nil", idx)
			failedSalePostRequests = append(failedSalePostRequests, FailedSalePostRequest{
				SaleID: "nil",
				Reason: err.Error(),
			})
		}
	}
	p.Wait()
}

func checkedBeforePosting(sales []vend.Sale9) {
	fmt.Println("\nPosting Sales..")

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(sales), "Posting sales")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	for idx, sale := range sales {
		bar.Increment()
		if sale.ID != nil {
			saleID := *sale.ID

			exists, err := saleExists(saleID)
			if err != nil {
				err = fmt.Errorf("error checking if sale exists: %s", err)
				failedSalePostRequests = append(failedSalePostRequests, FailedSalePostRequest{
					SaleID: saleID,
					Reason: err.Error(),
				})
				continue
			}

			if exists {
				reason := "sale has already been posted. Used mode overwrite true if you want to post anyway"
				failedSalePostRequests = append(failedSalePostRequests, FailedSalePostRequest{
					SaleID: saleID,
					Reason: reason,
				})
			} else {
				err = postSale(sale)
				if err != nil {
					err = fmt.Errorf("error posting sale: %s", err)
					failedSalePostRequests = append(failedSalePostRequests, FailedSalePostRequest{
						SaleID: saleID,
						Reason: err.Error(),
					})
					continue
				}
			}
		} else {
			err = fmt.Errorf("sale ID of sale number: %v is nil", idx)
			failedSalePostRequests = append(failedSalePostRequests, FailedSalePostRequest{
				SaleID: "nil",
				Reason: err.Error(),
			})
		}
	}
	p.Wait()
}

func postSale(sale vend.Sale9) error {

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

	fileName := fmt.Sprintf("%s_errored_sales_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		err = fmt.Errorf("error creating CSV file: %s", err)
		messenger.ExitWithError(err)
	}

	writer := csv.NewWriter(file)

	var warningMessage []string
	warningMessage = append(warningMessage, "WARNING:")
	warningMessage = append(warningMessage, "This CSV report is generated solely from the provided JSON data.")
	warningMessage = append(warningMessage, "It does not reflect the current status of sales")
	warningMessage = append(warningMessage, "Use this data for reference purposes only, and be aware that it may not accurately represent the latest sales information or its posting status in the system.")

	writer.Write(warningMessage)
	writer.Flush()

	return file, err
}

func GetVendDataForSalesReport(vc vend.Client) ([]vend.Register, []vend.User, []vend.Customer, map[string]string, []vend.Product) {
	// create a waitgroup to wait for all goroutines to finish
	fmt.Println("\nFetching data from Vend...")
	p, err := pbar.CreateMultiBarGroup(5, Token, DomainPrefix)
	if err != nil {
		fmt.Println("error creating progress bar group: ", err)
	}
	// create a channel to receive the results for each goroutine
	p.FetchDataWithProgressBar("registers")
	p.FetchDataWithProgressBar("users")
	p.FetchDataWithProgressBar("customers")
	p.FetchDataWithProgressBar("customerGroups")
	p.FetchDataWithProgressBar("products")

	p.MultiBarGroupWait()

	var registers []vend.Register
	var users []vend.User
	var customers []vend.Customer
	customerGroupMap := make(map[string]string)
	var products []vend.Product

	for err = range p.ErrorChannel {
		err = fmt.Errorf("error fetching data: %v", err)
		messenger.ExitWithError(err)
	}

	for data := range p.DataChannel {
		switch d := data.(type) {
		case []vend.Register:
			registers = d
		case []vend.User:
			users = d
		case []vend.Customer:
			customers = d
		case map[string]string:
			customerGroupMap = d
		case []vend.Product:
			products = d
		}
	}

	return registers, users, customers, customerGroupMap, products
}

func validateModeFlag(m string) bool {
	m = strings.ToLower(m)
	if m == "parse" {
		return false
	} else if m == "post" {
		return true
	} else {
		err := fmt.Errorf("'%s' is not a valid option for -m mode. Mode should be 'parse' or 'post'", m)
		messenger.ExitWithError(err)
	}
	return false
}

func prettyOverwriteBool(overwrite bool) string {
	if overwrite {
		return color.RedString("TRUE")
	} else {
		return color.YellowString("FALSE")
	}
}
