package config

import (
	"flag"
	"os"
)

type AppConfig struct {
	ServerBaseURL   string
	RedirectBaseURL string
}

var Config AppConfig

var defaultServerBaseURL = "localhost:8080"
var defaultRedirectBaseURL = "http://localhost:8080"

func InitConfiguration() {

	flag.StringVar(&Config.ServerBaseURL, "a", defaultServerBaseURL, "shortener server URL")
	flag.StringVar(&Config.RedirectBaseURL, "b", defaultRedirectBaseURL, "shortener redirect URL")

	flag.Parse()

	envServerBaseURL := os.Getenv("SERVER_ADDRESS")
	envRedirectBaseURL := os.Getenv("BASE_URL")

	if envServerBaseURL != "" {
		Config.ServerBaseURL = envServerBaseURL
	}

	if envRedirectBaseURL != "" {
		Config.RedirectBaseURL = envRedirectBaseURL
	}

}
