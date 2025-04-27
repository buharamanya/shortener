package config

import (
	"flag"
)

type AppConfig struct {
	ServerBaseURL   string
	RedirectBaseURL string
}

var Config AppConfig

func InitConfiguration() {

	flag.StringVar(&Config.ServerBaseURL, "a", "localhost:8080", "shortener server URL")
	flag.StringVar(&Config.RedirectBaseURL, "b", "http://localhost:8080", "shortener redirect URL")

	flag.Parse()

}
