package cmd

import (
	"fmt"

	"github.com/fatih/color"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vend/govend/vend"
	"github.com/vend/vend-cli/pkg/messenger"
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
		messenger.ExitWithError(err)
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
			messenger.ExitWithError(err)
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
