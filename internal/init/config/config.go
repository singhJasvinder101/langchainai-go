package config

import (
	"log"
	"os"
	"sync"

	"github.com/singhJasvinder101/cursor-go/internal/types"
	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "configs/config.yaml"

var cfg *types.Config
var once sync.Once

func MustInit(path string) {
	once.Do(func() {
		initConfig(path)
	})
}

func initConfig(path string) {
	if _, err := os.Stat(path); err != nil {
		log.Fatalf("config file not found: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	c := make(types.Config)
	if err := yaml.Unmarshal(raw, &c); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	cfg = &c
}

func GetConfig() *types.Config {
	return cfg
}

func GetString(path string) string {
	if cfg == nil {
		return ""
	}
	return cfg.GetString(path)
}

func GetInt(path string) int {
	if cfg == nil {
		return 0
	}
	return cfg.GetInt(path)
}

func GetBool(path string) bool {
	if cfg == nil {
		return false
	}
	return cfg.GetBool(path)
}
