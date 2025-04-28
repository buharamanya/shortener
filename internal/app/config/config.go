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

	Config.ServerBaseURL = os.Getenv("SERVER_ADDRESS")
	Config.RedirectBaseURL = os.Getenv("BASE_URL")

	if Config.ServerBaseURL == "" {
		flag.StringVar(&Config.ServerBaseURL, "a", defaultServerBaseURL, "shortener server URL")
	}

	if Config.RedirectBaseURL == "" {
		flag.StringVar(&Config.RedirectBaseURL, "b", defaultRedirectBaseURL, "shortener redirect URL")
	}

	flag.Parse()

}
