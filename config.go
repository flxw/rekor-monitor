package main

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type _Config struct {
	Database struct {
		Username string `yaml:"user" env:"DB_USERNAME" env-default:""`
		Password string `yaml:"password" env:"DB_PASSWORD" env-default:""`
		Name     string `yaml:"name" env:"DB_NAME" env-default:""`
		Url      string `yaml:"url" env:"DB_URL" env-default:""`
	} `yaml:"database"`
	Rekor struct {
		Url        string `yaml:"url" env:"REKOR_URL" env-default:"https://rekor.sigstore.dev"`
		UserAgent  string `yaml:"user-agent" env:"REKOR_CLIENT_USERAGENT" env-default:"monitor-crawler"`
		RetryCount uint   `yaml:"retry-count" env:"REKOR_CLIENT_RETRYCOUNT" env-default:"10"`
		StartIndex int64  `yaml:"initial-start-index" env:"REKOR_START_INDEX"`
	} `yaml:"rekor"`
}

var Config _Config

func init() {
	err := cleanenv.ReadConfig("config.yml", &Config)
	if err != nil {
		log.Fatal("Failed to read configuration", err)
	}
}
