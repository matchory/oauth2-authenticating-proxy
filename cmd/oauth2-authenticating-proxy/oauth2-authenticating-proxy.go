package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/matchory/oauth2-authenticating-proxy/internal/commands"
)

var configFile string
var rootCmd = &cobra.Command{
	Use:   "proxy",
	Short: "A reverse proxy to authenticate HTTP requests",
}

func CreateMainCommand() *cobra.Command {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file to use")

	rootCmd.AddCommand(commands.Serve)
	rootCmd.AddCommand(commands.Version)

	return rootCmd
}

func report(msg interface{}) {
	fmt.Printf("Error: %s\n", msg)
	os.Exit(1)
}

func initConfig() {
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/oauth2-proxy")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	}

	if err := viper.ReadInConfig(); err == nil {
		fmt.Printf("Using config file at '%s'\n", viper.ConfigFileUsed())
	}
}

func main() {
	err := CreateMainCommand().Execute()

	if err != nil {
		report(err)
	}
}
