package configuration

import (
	"encoding/json"
	"fmt"
	"github.com/labstack/gommon/log"
	"os"
	"path/filepath"
)

const (
	Development = "development"
	Staging     = "staging"
	Production  = "production"
)

type AppConfig struct {
	ConnectionStrings struct {
		Stocks string `json:"stocksDb"`
	}
	ApiSettings struct {
		Key string `json:"key"`
	}
}

var Configuration AppConfig

// set up base settings to prep configuration

// SetEnvironmentSettings map json configuration to our connection string struct
func SetEnvironmentSettings(env string) error {
	//get base config settings
	configBasePath, err := os.Getwd()
	if err != nil {
		return err
	}

	configBasePath = filepath.Join(configBasePath, "settings", "configuration")
	errConf := configureApp(configBasePath, env)
	if errConf != nil {
		log.Fatalf("Error setting environment settings: %s", errConf)
	}
	return nil
}

// configureApp open correct configuration based on environment
func configureApp(filePath string, env string) error {

	//open base configuration settings unless a specific environment is specified
	baseFile, err := os.Open(filePath + "\\config.json")
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
		return err
	}
	defer baseFile.Close()

	decoder := json.NewDecoder(baseFile)

	baseConfig := AppConfig{}

	if err = decoder.Decode(&baseConfig); err != nil {
		log.Fatalf("Error Decoding Configuration File: %v", err)
		return err
	}

	currentEnv := fmt.Sprintf("\\config.%s.json", env)
	envFilePath := filepath.Join(filePath, currentEnv)

	envFile, err := os.Open(envFilePath)
	if err != nil {
		log.Fatalf("Error opening %s config file: %v", env, err)
		return err
	}
	defer envFile.Close()

	decoder = json.NewDecoder(envFile)
	envConfig := AppConfig{}

	if err = decoder.Decode(&envConfig); err != nil {
		log.Fatalf("Error Decoding Environment Configuration File: %v", err)
	}

	baseConfig.ConnectionStrings = envConfig.ConnectionStrings
	baseConfig.ApiSettings = envConfig.ApiSettings

	Configuration = baseConfig

	return nil
}
