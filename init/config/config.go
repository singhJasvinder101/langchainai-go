package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/singhJasvinder101/ai-go/pkg/log"
	"github.com/singhJasvinder101/ai-go/internal/types"
	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "configs/config.yaml"

var (
	cfg  *types.Config
	once sync.Once
)

func MustInit(path string) {
	if len(path) == 0 {
		path = DefaultConfigPath
	}
	once.Do(func() {
		if err := initConfig(path); err != nil {
			log.Fatal("config initialization failed", "error", err, "path", path)
		}
	})
}

func initConfig(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("config file not found: %w", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	c := make(types.Config)
	if err := yaml.Unmarshal(raw, &c); err != nil {
		return fmt.Errorf("parse config: %w", err)
	}

	cfg = &c
	return nil
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
