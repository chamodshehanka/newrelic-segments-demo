package configs

import (
	"flag"
	"log"
	"sync"

	"github.com/spf13/viper"
)

var (
	config *Config
	once   sync.Once
)

// LoadConfig loads YAML config using Viper and preloads JSON files
func LoadConfig(configFile string) (*Config, error) {
	viper.SetConfigFile(configFile)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file: %v", err)
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Printf("Error unmarshalling config: %v", err)
		return nil, err
	}

	log.Printf("Config loaded successfully")
	return &config, nil
}

// GetConfig loads the config once, accepting a file path via command-line flag `-config`
func GetConfig() *Config {
	once.Do(func() {
		// Parse the command-line flag for config file path
		configFile := flag.String("config", "", "Path to the configuration file")
		flag.Parse()

		if *configFile == "" {
			configFile = new(string) // Ensure it's not nil
			*configFile = "./config.yaml"
		}

		log.Printf("Using config: %s", *configFile)

		var err error
		config, err = LoadConfig(*configFile)
		if err != nil {
			log.Fatalf("Error loading config: %v", err)
		}
	})
	return config
}
