package config

import (
	"flag"
	"os"
)

const (
	defaultServerBaseURL   = "localhost:8080"
	defaultRedirectBaseURL = "http://localhost:8080"
	defaultStorageFileName = "storage.txt"
	defaultDataBaseDSN     = "" // поменять на следующих итерациях
	defaultSecretKey       = "secret"
)

type AppConfig struct {
	ServerBaseURL   string
	RedirectBaseURL string
	StorageFileName string
	DataBaseDSN     string
	SecretKey       string
}

var AppParams AppConfig

func InitConfiguration() *AppConfig {

	flag.StringVar(&AppParams.ServerBaseURL, "a", defaultServerBaseURL, "shortener server URL")
	flag.StringVar(&AppParams.RedirectBaseURL, "b", defaultRedirectBaseURL, "shortener redirect URL")
	flag.StringVar(&AppParams.StorageFileName, "f", defaultStorageFileName, "shortener storage filename")
	flag.StringVar(&AppParams.DataBaseDSN, "d", defaultDataBaseDSN, "shortener database DSN")
	flag.StringVar(&AppParams.SecretKey, "s", defaultSecretKey, "key for JWT")

	flag.Parse()

	envServerBaseURL := os.Getenv("SERVER_ADDRESS")
	envRedirectBaseURL := os.Getenv("BASE_URL")
	envStorageFileName := os.Getenv("FILE_STORAGE_PATH")
	envDataBaseDSN := os.Getenv("DATABASE_DSN")
	envSecretKey := os.Getenv("SECRET_KEY")

	if envServerBaseURL != "" {
		AppParams.ServerBaseURL = envServerBaseURL
	}

	if envRedirectBaseURL != "" {
		AppParams.RedirectBaseURL = envRedirectBaseURL
	}

	if envStorageFileName != "" {
		AppParams.StorageFileName = envStorageFileName
	}

	if envDataBaseDSN != "" {
		AppParams.DataBaseDSN = envDataBaseDSN
	}

	if envSecretKey != "" {
		AppParams.SecretKey = envSecretKey
	}

	return &AppParams

}
