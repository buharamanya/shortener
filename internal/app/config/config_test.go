package config

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

// Вспомогательная функция для сброса состояния флагов и переменных окружения
func reset() {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError) // Сбрасываем флаги
	os.Unsetenv("SERVER_ADDRESS")                                        // Очищаем переменные окружения
	os.Unsetenv("BASE_URL")
	os.Unsetenv("FILE_STORAGE_PATH")
	os.Unsetenv("DATABASE_DSN")
	os.Unsetenv("SECRET_KEY")
	os.Unsetenv("ENABLE_HTTPS")
}

func TestInitConfiguration_DefaultValues(t *testing.T) {
	reset()

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
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка установки значений по умолчанию. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_CommandLineFlags(t *testing.T) {
	reset()

	// Сохраняем оригинальные аргументы и восстанавливаем их после теста
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-a=flag:8080", "-b=http://flag:8080", "-f=flag.txt", "-d=flag_DATABASE_DSN", "-k=flag_key", "-s=true"}

	config := InitConfiguration()

	expected := &AppConfig{
		ServerBaseURL:   "flag:8080",
		RedirectBaseURL: "http://flag:8080",
		StorageFileName: "flag.txt",
		DataBaseDSN:     "flag_DATABASE_DSN",
		SecretKey:       "flag_key",
		EnableHTTPS:     true,
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка установки переменных командной строки. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_EnvironmentVariables(t *testing.T) {
	reset()

	// Устанавливаем переменные окружения
	os.Setenv("SERVER_ADDRESS", "env:8080")
	os.Setenv("BASE_URL", "http://env:8080")
	os.Setenv("FILE_STORAGE_PATH", "env.txt")
	os.Setenv("DATABASE_DSN", "env_DATABASE_DSN")
	os.Setenv("SECRET_KEY", "env_SECRET_KEY")
	os.Setenv("ENABLE_HTTPS", "true")

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
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка установки переменных среды. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_Priority(t *testing.T) {
	reset()

	// Устанавливаем и флаги, и переменные окружения
	os.Setenv("SERVER_ADDRESS", "env:8080")
	os.Setenv("BASE_URL", "http://env:8080")
	os.Setenv("FILE_STORAGE_PATH", "env.txt")
	os.Setenv("DATABASE_DSN", "env_DATABASE_DSN")
	os.Setenv("SECRET_KEY", "env_SECRET_KEY")
	os.Setenv("ENABLE_HTTPS", "true")

	// Сохраняем оригинальные аргументы и восстанавливаем их после теста
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-a=flag:8080", "-b=http://flag:8080", "-f=flag.txt", "-d=flag_DATABASE_DSN", "-k=flag_key", "-s=false"}

	config := InitConfiguration()

	// Ожидаем, что переменные окружения имеют приоритет над флагами командной строки
	expected := &AppConfig{
		ServerBaseURL:   "env:8080",
		RedirectBaseURL: "http://env:8080",
		StorageFileName: "env.txt",
		DataBaseDSN:     "env_DATABASE_DSN",
		SecretKey:       "env_SECRET_KEY",
		EnableHTTPS:     true, // Переменная окружения имеет приоритет
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка приоритета выбора переменных. Ожидалось %v, получено %v", expected, config)
	}
}

func TestInitConfiguration_EnableHTTPS_EnvironmentValues(t *testing.T) {
	reset()

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
			reset()
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
