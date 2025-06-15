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
	os.Args = []string{"cmd", "-a=flag:8080", "-b=http://flag:8080", "-f=flag.txt", "-d=flag_DATABASE_DSN"}

	config := InitConfiguration()

	expected := &AppConfig{
		ServerBaseURL:   "flag:8080",
		RedirectBaseURL: "http://flag:8080",
		StorageFileName: "flag.txt",
		DataBaseDSN:     "flag_DATABASE_DSN",
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

	// Сохраняем оригинальные аргументы и восстанавливаем их после теста
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "-a=flag:8080", "-b=http://flag:8080", "-f=flag.txt", "-d=flag_DATABASE_DSN"}

	config := InitConfiguration()

	// Ожидаем, что переменные окружения имеют приоритет над флагами командной строки
	expected := &AppConfig{
		ServerBaseURL:   "env:8080",
		RedirectBaseURL: "http://env:8080",
		StorageFileName: "env.txt",
		DataBaseDSN:     "env_DATABASE_DSN",
	}

	if !reflect.DeepEqual(config, expected) {
		t.Errorf("Ошибка приоритета выбора переменных. Ожидалось %v, получено %v", expected, config)
	}
}
