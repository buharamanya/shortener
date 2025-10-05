package config

import (
	"encoding/json"
	"flag"
	"os"
	"reflect"
	"testing"
)

// Вспомогательная функция для сброса состояния флагов и переменных окружения
func reset(t *testing.T) {
	t.Helper()

	// Сбрасываем флаги путем создания нового FlagSet
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Очищаем глобальную переменную конфигурации
	AppParams = AppConfig{}

	// Очищаем переменные окружения
	os.Unsetenv("SERVER_ADDRESS")
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("ENABLE_HTTPS")
	os.Unsetenv("TRUSTED_SUBNET")
	os.Unsetenv("GRPC_PORT")
	os.Unsetenv("CONFIG")
}

func TestInitConfiguration_DefaultValues(t *testing.T) {
	reset(t)

	// Сохраняем оригинальные аргументы и восстанавливаем их после теста
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"} // Пустые аргументы командной строки

	config := InitConfiguration()

	expected := &AppConfig{
		ServerBaseURL:   defaultServerBaseURL,
		RedirectBaseURL: defaultRedirectBaseURL,
		StorageFileName: defaultStorageFileName,
		DataBaseDSN:     defaultDataBaseDSN,
		SecretKey:       defaultSecretKey,
		EnableHTTPS:     defaultEnableHTTPS,
		TrustedSubnet:   defaultTrustedSubnet,
		GRPCPort:        defaultGRPCPort,
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка установки значений по умолчанию. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_CommandLineFlags(t *testing.T) {
	reset(t)

	// Сохраняем оригинальные аргументы и восстанавливаем их после теста
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-a=flag:8080", "-b=http://flag:8080", "-f=flag.txt", "-d=flag_DATABASE_DSN", "-k=flag_key", "-s=true", "-t=192.168.1.0/24", "-grpc-port=9090"}

	config := InitConfiguration()

	expected := &AppConfig{
		ServerBaseURL:   "flag:8080",
		RedirectBaseURL: "http://flag:8080",
		StorageFileName: "flag.txt",
		DataBaseDSN:     "flag_DATABASE_DSN",
		SecretKey:       "flag_key",
		EnableHTTPS:     true,
		TrustedSubnet:   "192.168.1.0/24",
		GRPCPort:        "9090",
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка установки переменных командной строки. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_EnvironmentVariables(t *testing.T) {
	reset(t)

	// Устанавливаем переменные окружения
	os.Setenv("SERVER_ADDRESS", "env:8080")
	os.Setenv("BASE_URL", "http://env:8080")
	os.Setenv("FILE_STORAGE_PATH", "env.txt")
	os.Setenv("DATABASE_DSN", "env_DATABASE_DSN")
	os.Setenv("SECRET_KEY", "env_SECRET_KEY")
	os.Setenv("ENABLE_HTTPS", "true")
	os.Setenv("TRUSTED_SUBNET", "10.0.0.0/24")
	os.Setenv("GRPC_PORT", "50052")

	// Пустые аргументы командной строки
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	config := InitConfiguration()

	expected := &AppConfig{
		ServerBaseURL:   "env:8080",
		RedirectBaseURL: "http://env:8080",
		StorageFileName: "env.txt",
		DataBaseDSN:     "env_DATABASE_DSN",
		SecretKey:       "env_SECRET_KEY",
		EnableHTTPS:     true,
		TrustedSubnet:   "10.0.0.0/24",
		GRPCPort:        "50052",
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка установки переменных среды. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_Priority(t *testing.T) {
	reset(t)

	// Устанавливаем и флаги, и переменные окружения
	os.Setenv("SERVER_ADDRESS", "env:8080")
	os.Setenv("BASE_URL", "http://env:8080")
	os.Setenv("FILE_STORAGE_PATH", "env.txt")
	os.Setenv("DATABASE_DSN", "env_DATABASE_DSN")
	os.Setenv("SECRET_KEY", "env_SECRET_KEY")
	os.Setenv("ENABLE_HTTPS", "true")
	os.Setenv("TRUSTED_SUBNET", "10.0.0.0/24")
	os.Setenv("GRPC_PORT", "50052")

	// Сохраняем оригинальные аргументы и восстанавливаем их после теста
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-a=flag:8080", "-b=http://flag:8080", "-f=flag.txt", "-d=flag_DATABASE_DSN", "-k=flag_key", "-s=false", "-t=192.168.1.0/24", "-grpc-port=9090"}

	config := InitConfiguration()

	// Ожидаем, что переменные окружения имеют приоритет над флагами командной строки
	expected := &AppConfig{
		ServerBaseURL:   "env:8080",
		RedirectBaseURL: "http://env:8080",
		StorageFileName: "env.txt",
		DataBaseDSN:     "env_DATABASE_DSN",
		SecretKey:       "env_SECRET_KEY",
		EnableHTTPS:     true, // Переменная окружения имеет приоритет
		TrustedSubnet:   "10.0.0.0/24",
		GRPCPort:        "50052",
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка приоритета выбора переменных. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_EnableHTTPS_EnvironmentValues(t *testing.T) {
	reset(t)

	testCases := []struct {
		name     string
		envValue string
		expected bool
	}{
		{"true value", "true", true},
		{"1 value", "1", true},
		{"false value", "false", false},
		{"empty value", "", false},
		{"other value", "yes", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reset(t)
			os.Setenv("ENABLE_HTTPS", tc.envValue)

			oldArgs := os.Args
			defer func() { os.Args = oldArgs }()
			os.Args = []string{"cmd"}

			config := InitConfiguration()

			if config.EnableHTTPS != tc.expected {
				t.Errorf("Для значения '%s' ожидалось %v, получено %v", tc.envValue, tc.expected, config.EnableHTTPS)
			}
		})
	}
}

func TestInitConfiguration_ConfigFile(t *testing.T) {
	reset(t)

	// Создаем временный конфигурационный файл
	configData := AppConfig{
		ServerBaseURL:   "file:8080",
		RedirectBaseURL: "http://file:8080",
		StorageFileName: "file.txt",
		DataBaseDSN:     "file_DATABASE_DSN",
		SecretKey:       "file_SECRET_KEY",
		EnableHTTPS:     true,
		TrustedSubnet:   "172.16.0.0/24",
		GRPCPort:        "50053",
	}

	tempFile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatal("Cannot create temp file:", err)
	}
	defer os.Remove(tempFile.Name())

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(configData); err != nil {
		t.Fatal("Cannot encode config to file:", err)
	}
	tempFile.Close()

	// Устанавливаем переменную окружения CONFIG
	os.Setenv("CONFIG", tempFile.Name())

	// Пустые аргументы командной строки
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	config := InitConfiguration()

	if !reflect.DeepEqual(config, &configData) {
		t.Errorf("Ошибка загрузки конфигурации из файла. Ожидалось %v, получено %v", configData, config)
	}
}

func TestInitConfiguration_ConfigFilePriority(t *testing.T) {
	reset(t)

	// Создаем временный конфигурационный файл
	configData := AppConfig{
		ServerBaseURL:   "file:8080",
		RedirectBaseURL: "http://file:8080",
		StorageFileName: "file.txt",
		DataBaseDSN:     "file_DATABASE_DSN",
		SecretKey:       "file_SECRET_KEY",
		EnableHTTPS:     true,
		TrustedSubnet:   "172.16.0.0/24",
		GRPCPort:        "50053",
	}

	tempFile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatal("Cannot create temp file:", err)
	}
	defer os.Remove(tempFile.Name())

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(configData); err != nil {
		t.Fatal("Cannot encode config to file:", err)
	}
	tempFile.Close()

	// Устанавливаем переменные окружения (должны иметь приоритет над файлом)
	os.Setenv("SERVER_ADDRESS", "env:8080")
	os.Setenv("BASE_URL", "http://env:8080")
	os.Setenv("CONFIG", tempFile.Name())

	// Пустые аргументы командной строки
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd"}

	config := InitConfiguration()

	expected := &AppConfig{
		ServerBaseURL:   "env:8080",          // из env, а не из файла
		RedirectBaseURL: "http://env:8080",   // из env, а не из файла
		StorageFileName: "file.txt",          // из файла
		DataBaseDSN:     "file_DATABASE_DSN", // из файла
		SecretKey:       "file_SECRET_KEY",   // из файла
		EnableHTTPS:     true,                // из файла
		TrustedSubnet:   "172.16.0.0/24",     // из файла
		GRPCPort:        "50053",             // из файла
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка приоритета конфигурации. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_ConfigFileFlag(t *testing.T) {
	reset(t)

	// Создаем временный конфигурационный файл
	configData := AppConfig{
		ServerBaseURL:   "file:8080",
		RedirectBaseURL: "http://file:8080",
		StorageFileName: "file.txt",
		DataBaseDSN:     "file_DATABASE_DSN",
		SecretKey:       "file_SECRET_KEY",
		EnableHTTPS:     true,
		TrustedSubnet:   "172.16.0.0/24",
		GRPCPort:        "50053",
	}

	tempFile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatal("Cannot create temp file:", err)
	}
	defer os.Remove(tempFile.Name())

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(configData); err != nil {
		t.Fatal("Cannot encode config to file:", err)
	}
	tempFile.Close()

	// Используем флаг -c вместо переменной окружения
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-c=" + tempFile.Name()}

	config := InitConfiguration()

	if !reflect.DeepEqual(config, &configData) {
		t.Errorf("Ошибка загрузки конфигурации из файла через флаг. Ожидалось %v, получено %v", configData, config)
	}
}
