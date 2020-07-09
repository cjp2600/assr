package cmd

import (
	"fmt"
	"github.com/cjp2600/assr/log"
	"github.com/spf13/cobra"
	"os"

	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "assr",
	Short: "ass-server-side rendering",
	Long:  ``,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringP("config", "c", "", "Config file assr.yaml")
}

func initConfig() {
	cfgFile, err := rootCmd.Flags().GetString("config")
	if err != nil {
		log.Fatal(err.Error())
	}
	viper.SetConfigFile(cfgFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		log.Printf("Using config file: %s \n", viper.ConfigFileUsed())
	} else {
		log.Error(fmt.Errorf("config file not found"))
		os.Exit(0)
	}
}
