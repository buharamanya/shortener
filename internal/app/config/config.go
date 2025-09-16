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
	defaultEnableHTTPS     = false
)

// структура для конфига.
type AppConfig struct {
	ServerBaseURL   string
	RedirectBaseURL string
	StorageFileName string
	DataBaseDSN     string
	SecretKey       string
	EnableHTTPS     bool
}

// глобальный конфиг.
var AppParams AppConfig

// инит глобального конфига.
func InitConfiguration() *AppConfig {

	flag.StringVar(&AppParams.ServerBaseURL, "a", defaultServerBaseURL, "shortener server URL")
	flag.StringVar(&AppParams.RedirectBaseURL, "b", defaultRedirectBaseURL, "shortener redirect URL")
	flag.StringVar(&AppParams.StorageFileName, "f", defaultStorageFileName, "shortener storage filename")
	flag.StringVar(&AppParams.DataBaseDSN, "d", defaultDataBaseDSN, "shortener database DSN")
	flag.StringVar(&AppParams.SecretKey, "k", defaultSecretKey, "key for JWT")
	flag.BoolVar(&AppParams.EnableHTTPS, "s", defaultEnableHTTPS, "enable HTTPS")

	flag.Parse()

	envServerBaseURL := os.Getenv("SERVER_ADDRESS")
	envRedirectBaseURL := os.Getenv("BASE_URL")
	envStorageFileName := os.Getenv("FILE_STORAGE_PATH")
	envDataBaseDSN := os.Getenv("DATABASE_DSN")
	envSecretKey := os.Getenv("SECRET_KEY")
	envEnableHTTPS := os.Getenv("ENABLE_HTTPS")

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

	if envEnableHTTPS == "true" || envEnableHTTPS == "1" {
		AppParams.EnableHTTPS = true
	}

	return &AppParams

}
