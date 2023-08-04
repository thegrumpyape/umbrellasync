package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	logPath string
	logFile *os.File
	logger  *log.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "umbrellasync",
	Short: "Syncs blocklists between files and Umbrella",
	Long:  ``,
}

func Execute() {
	if logFile != nil {
		defer logFile.Close()
	}

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	if logPath == "" {
		logPath = "umbrellasync.log"
	}

	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	logger = log.New(os.Stdout, "umbrella: ", log.LstdFlags)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.umbrellasync.yaml)")
	rootCmd.PersistentFlags().StringVar(&logPath, "log", "", "log file (default is ./umbrellasync.log)")

	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".umbrellasync" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".umbrellasync")

		cfgFile = fmt.Sprintf("%s/.umbrellasync.yaml", home)
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; Create config file
			fmt.Println("No config file found at:", cfgFile)
			err := CreateConfig()
			if err != nil {
				log.Fatal(err)
			}
			viper.SetConfigFile(cfgFile)

		} else {
			// Config file was found but another error was produced
			fmt.Println("Error reading config file:", viper.ConfigFileUsed())
			fmt.Println("Check the config file formatting")
			log.Fatal(err)
		}
	}
}
