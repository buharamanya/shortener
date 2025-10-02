package config

import (
	"encoding/json"
	"flag"
	"os"

	"github.com/buharamanya/shortener/internal/app/logger"
	"go.uber.org/zap"
)

const (
	defaultServerBaseURL   = "localhost:8080"
	defaultRedirectBaseURL = "http://localhost:8080"
	defaultStorageFileName = "storage.txt"
	defaultDataBaseDSN     = "" // поменять на следующих итерациях
	defaultSecretKey       = "secret"
	defaultEnableHTTPS     = false
	defaultConfigFile      = ""
)

// структура для конфига.
type AppConfig struct {
	ServerBaseURL   string `json:"server_address"`
	RedirectBaseURL string `json:"base_url"`
	StorageFileName string `json:"file_storage_path"`
	DataBaseDSN     string `json:"database_dsn"`
	SecretKey       string `json:"secret_key"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// глобальный конфиг.
var AppParams AppConfig

// инит глобального конфига.
func InitConfiguration() *AppConfig {
	var configFile string

	flag.StringVar(&AppParams.ServerBaseURL, "a", defaultServerBaseURL, "shortener server URL")
	flag.StringVar(&AppParams.RedirectBaseURL, "b", defaultRedirectBaseURL, "shortener redirect URL")
	flag.StringVar(&AppParams.StorageFileName, "f", defaultStorageFileName, "shortener storage filename")
	flag.StringVar(&AppParams.DataBaseDSN, "d", defaultDataBaseDSN, "shortener database DSN")
	flag.StringVar(&AppParams.SecretKey, "k", defaultSecretKey, "key for JWT")
	flag.BoolVar(&AppParams.EnableHTTPS, "s", defaultEnableHTTPS, "enable HTTPS")
	flag.StringVar(&configFile, "c", defaultConfigFile, "config file path")

	flag.Parse()

	// Загружаем конфигурацию из файла, если указан
	if configFile == "" {
		configFile = os.Getenv("CONFIG")
	}

	if configFile != "" {
		loadConfigFromFile(configFile)
	}

	// Переменные окружения имеют приоритет над файлом конфигурации
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

// loadConfigFromFile загружает конфигурацию из JSON файла
func loadConfigFromFile(filename string) {
	file, err := os.Open(filename)
	if err != nil {
		logger.Log.Error("Warning: Cannot open config file ", zap.String("file", filename))
		return
	}
	defer file.Close()

	var fileConfig AppConfig
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&fileConfig); err != nil {
		logger.Log.Error("Warning: Cannot open config file ", zap.String("file", filename))
		return
	}

	// Устанавливаем значения из файла только если они не пустые
	if fileConfig.ServerBaseURL != "" {
		AppParams.ServerBaseURL = fileConfig.ServerBaseURL
	}
	if fileConfig.RedirectBaseURL != "" {
		AppParams.RedirectBaseURL = fileConfig.RedirectBaseURL
	}
	if fileConfig.StorageFileName != "" {
		AppParams.StorageFileName = fileConfig.StorageFileName
	}
	if fileConfig.DataBaseDSN != "" {
		AppParams.DataBaseDSN = fileConfig.DataBaseDSN
	}
	if fileConfig.SecretKey != "" {
		AppParams.SecretKey = fileConfig.SecretKey
	}
	// Для булевых значений проверяем, было ли значение установлено в файле
	// (в Go булево значение по умолчанию false, поэтому используем отдельный флаг)
	if fileConfig.EnableHTTPS {
		AppParams.EnableHTTPS = fileConfig.EnableHTTPS
	}
}
