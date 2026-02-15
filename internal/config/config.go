package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	LLM             LLMConfig `mapstructure:"llm" yaml:"llm"`
	DefaultLanguage string    `mapstructure:"default_language" yaml:"default_language"`
	Verbose         bool      `mapstructure:"verbose" yaml:"verbose"`
}

type LLMConfig struct {
	Provider  string   `mapstructure:"provider" yaml:"provider"`
	Model     string   `mapstructure:"model" yaml:"model"`
	APIKey    string   `mapstructure:"api_key" yaml:"api_key"`
	BaseURL   string   `mapstructure:"base_url" yaml:"base_url"`
	Fallbacks []string `mapstructure:"fallbacks" yaml:"fallbacks"`
}

func defaultConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}
	return filepath.Join(home, ".archway", "config.yaml"), nil
}

func Load(configPath string) (*AppConfig, error) {
	path := strings.TrimSpace(configPath)
	if path == "" {
		var err error
		path, err = defaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetDefault("default_language", "go")
	v.SetDefault("verbose", false)
	v.SetDefault("llm.provider", "")
	v.SetDefault("llm.model", "")
	v.SetDefault("llm.api_key", "")
	v.SetDefault("llm.base_url", "")
	v.SetDefault("llm.fallbacks", []string{})

	v.SetEnvPrefix("ARCHWAY")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var cfgNotFound viper.ConfigFileNotFoundError
		if !errors.As(err, &cfgNotFound) && !os.IsNotExist(err) {
			return nil, fmt.Errorf("read config file %s: %w", path, err)
		}
	}

	cfg := &AppConfig{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	return cfg, nil
}

func Save(configPath string, cfg *AppConfig) error {
	if cfg == nil {
		return fmt.Errorf("config is nil")
	}

	path := strings.TrimSpace(configPath)
	if path == "" {
		var err error
		path, err = defaultConfigPath()
		if err != nil {
			return err
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	v := viper.New()
	v.SetConfigType("yaml")
	v.Set("default_language", cfg.DefaultLanguage)
	v.Set("verbose", cfg.Verbose)
	v.Set("llm.provider", cfg.LLM.Provider)
	v.Set("llm.model", cfg.LLM.Model)
	v.Set("llm.api_key", cfg.LLM.APIKey)
	v.Set("llm.base_url", cfg.LLM.BaseURL)
	v.Set("llm.fallbacks", cfg.LLM.Fallbacks)

	content, err := yaml.Marshal(v.AllSettings())
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(path, content, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}
	return nil
}
