package config

import (
	"flag"
	"os"
)

const (
	defaultServerBaseURL   = "localhost:8080"
	defaultRedirectBaseURL = "http://localhost:8080"
)

type AppConfig struct {
	ServerBaseURL   string
	RedirectBaseURL string
}

func InitConfiguration() *AppConfig {

	var config AppConfig

	flag.StringVar(&config.ServerBaseURL, "a", defaultServerBaseURL, "shortener server URL")
	flag.StringVar(&config.RedirectBaseURL, "b", defaultRedirectBaseURL, "shortener redirect URL")

	flag.Parse()

	envServerBaseURL := os.Getenv("SERVER_ADDRESS")
	envRedirectBaseURL := os.Getenv("BASE_URL")

	if envServerBaseURL != "" {
		config.ServerBaseURL = envServerBaseURL
	}

	if envRedirectBaseURL != "" {
		config.RedirectBaseURL = envRedirectBaseURL
	}

	return &config

}
