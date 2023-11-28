package cmd

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vend/govend/vend"
)

const version = "1.8"

// Variables for Client authentication details and flags
var (
	DomainPrefix string
	Token        string
	vendClient   *vend.Client
	FilePath     string
	cfgFile      string
	logo         = color.GreenString(`                             _ 
 __   __   ___   _ __     __| |
 \ \ / /  / _ \ | '_ \   / _  |
  \ V /  |  __/ | | | | | (_| |
   \_/    \___| |_| |_|  \__,_|`)
)

// Command Config
var rootCmd = &cobra.Command{
	Use:     "vendcli",
	Version: version,
	Short: fmt.Sprintf(`
%s`, logo)}

func init() {
	cobra.OnInitialize(initConfig)

	// Get store info from command line flags.
	rootCmd.PersistentFlags().StringVarP(&DomainPrefix, "Domain", "d", "", "The Vend store name (prefix in xxxx.vendhq.com)")
	rootCmd.PersistentFlags().StringVarP(&Token, "Token", "t", "", "API Access Token for the store, Setup -> Personal Tokens.")
	rootCmd.MarkPersistentFlagRequired("Domain")
	rootCmd.MarkPersistentFlagRequired("Token")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		panic(vend.Exit{1})
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			panic(vend.Exit{1})
		}

		// Search config in home directory with name ".vend" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".vendcli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

// Read passed CSV and returns the IDs
func readCSV(FilePath string) ([]string, error) {

	// Open our provided CSV file
	file, err := os.Open(FilePath)
	if err != nil {
		errorMsg := `error opening csv file - please check you've specified the right file

Tip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`
		fmt.Println(errorMsg, "\n")
		return nil, err
	}

	// Make sure to close the file
	defer file.Close()

	// Create CSV read on our file
	reader := csv.NewReader(file)

	// Read the rest of the data from the CSV
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	var rowNumber int
	entities := []string{}

	// Loop through rows and assign them
	for _, row := range rows {
		rowNumber++
		entityIDs := row[0]
		entities = append(entities, entityIDs)
	}

	return entities, err
}
