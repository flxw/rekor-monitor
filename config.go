package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type _Config struct {
	Database struct {
		String string `yaml:"string" env:"DATABASE_URL" env-default:"test:test@tcp(localhost)/test"`
	} `yaml:"database"`
	Rekor struct {
		Url        string `yaml:"url" env:"REKOR_URL" env-default:"https://rekor.sigstore.dev"`
		UserAgent  string `yaml:"user-agent" env:"REKOR_CLIENT_USERAGENT" env-default:"crawler/rekor-monitor.flxw.de"`
		RetryCount uint   `yaml:"retry-count" env:"REKOR_CLIENT_RETRYCOUNT" env-default:"10"`
		StartIndex int64  `yaml:"initial-start-index" env:"REKOR_START_INDEX"`
	} `yaml:"rekor"`
}

var Config _Config

func init() {
	err := cleanenv.ReadConfig("config.yml", &Config)
	if err != nil {
		log.Println("WARNING: Could not read the configuration file - falling back to defaults and environment variables")
	}
	cleanenv.ReadEnv(&Config)
}
