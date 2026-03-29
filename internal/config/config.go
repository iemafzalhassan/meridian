// Copyright 2026 Meridian OSS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	DefaultConfigDirName  = ".meridian"
	DefaultConfigFileName = "config.yaml"
)

type Config struct {
	LLM        LLMConfig        `mapstructure:"llm"`
	Embeddings EmbeddingsConfig `mapstructure:"embeddings"`
	Qdrant     QdrantConfig     `mapstructure:"qdrant"`
	GitHub     GitHubConfig     `mapstructure:"github"`
	Logging    LoggingConfig    `mapstructure:"logging"`
}

type LLMConfig struct {
	Provider      string        `mapstructure:"provider"`
	Model         string        `mapstructure:"model"`
	OllamaURL     string        `mapstructure:"ollama_url"`
	OpenRouterKey string        `mapstructure:"openrouter_key"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryCount    int           `mapstructure:"retry_count"`
}

type EmbeddingsConfig struct {
	Provider   string        `mapstructure:"provider"`
	Model      string        `mapstructure:"model"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryCount int           `mapstructure:"retry_count"`
}

type QdrantConfig struct {
	URL           string        `mapstructure:"url"`
	Collection    string        `mapstructure:"collection"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryCount    int           `mapstructure:"retry_count"`
	EmbeddingDims int           `mapstructure:"embedding_dims"`
}

type GitHubConfig struct {
	Token      string        `mapstructure:"token"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryCount int           `mapstructure:"retry_count"`
}

type LoggingConfig struct {
	Level string `mapstructure:"level"`
}

func Load() (*Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("config: resolve user home: %w", err)
	}

	configFile := filepath.Join(home, DefaultConfigDirName, DefaultConfigFileName)
	v := viper.New()
	v.SetConfigFile(configFile)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("MERIDIAN")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	_ = v.BindEnv("llm.openrouter_key", "MERIDIAN_OPENROUTER_KEY")
	_ = v.BindEnv("llm.ollama_url", "MERIDIAN_OLLAMA_URL")
	_ = v.BindEnv("github.token", "MERIDIAN_GITHUB_TOKEN")

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) && !os.IsNotExist(err) {
			return nil, fmt.Errorf("config: read config file: %w", err)
		}
	}

	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("config: unmarshal config: %w", err)
	}

	return cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("llm.provider", "ollama")
	v.SetDefault("llm.model", "llama3.2")
	v.SetDefault("llm.ollama_url", "http://localhost:11434")
	v.SetDefault("llm.openrouter_key", "")
	v.SetDefault("llm.timeout", 30*time.Second)
	v.SetDefault("llm.retry_count", 2)

	v.SetDefault("embeddings.provider", "ollama")
	v.SetDefault("embeddings.model", "nomic-embed-text")
	v.SetDefault("embeddings.timeout", 30*time.Second)
	v.SetDefault("embeddings.retry_count", 2)

	v.SetDefault("qdrant.url", "http://localhost:6334")
	v.SetDefault("qdrant.collection", "meridian_projects")
	v.SetDefault("qdrant.timeout", 30*time.Second)
	v.SetDefault("qdrant.retry_count", 2)
	v.SetDefault("qdrant.embedding_dims", 768)

	v.SetDefault("github.token", "")
	v.SetDefault("github.timeout", 30*time.Second)
	v.SetDefault("github.retry_count", 2)

	v.SetDefault("logging.level", "info")
}
