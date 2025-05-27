package config

import (
	"flag"
	"os"
)

const (
	defaultServerBaseURL   = "localhost:8080"
	defaultRedirectBaseURL = "http://localhost:8080"
	defaultStorageFileName = "storage.txt"
)

type AppConfig struct {
	ServerBaseURL   string
	RedirectBaseURL string
	StorageFileName string
}

func InitConfiguration() *AppConfig {

	var config AppConfig

	flag.StringVar(&config.ServerBaseURL, "a", defaultServerBaseURL, "shortener server URL")
	flag.StringVar(&config.RedirectBaseURL, "b", defaultRedirectBaseURL, "shortener redirect URL")
	flag.StringVar(&config.StorageFileName, "f", defaultStorageFileName, "shortener storage filename")

	flag.Parse()

	envServerBaseURL := os.Getenv("SERVER_ADDRESS")
	envRedirectBaseURL := os.Getenv("BASE_URL")
	envStorageFileName := os.Getenv("FILE_STORAGE_PATH")

	if envServerBaseURL != "" {
		config.ServerBaseURL = envServerBaseURL
	}

	if envRedirectBaseURL != "" {
		config.RedirectBaseURL = envRedirectBaseURL
	}

	if envStorageFileName != "" {
		config.StorageFileName = envStorageFileName
	}

	return &config

}
