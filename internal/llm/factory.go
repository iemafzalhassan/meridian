// Copyright 2026 Meridian OSS Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package llm

import (
	"fmt"
	"strings"

	"github.com/iemafzalhassan/meridian/internal/config"
)

// NewFromConfig returns the LLM provider for chat/completion (and optional embed on the same provider).
func NewFromConfig(cfg *config.Config) (Provider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm: missing config")
	}
	switch strings.ToLower(strings.TrimSpace(cfg.LLM.Provider)) {
	case "ollama":
		return NewOllamaProvider(cfg)
	case "openrouter":
		return NewOpenRouterProvider(cfg)
	default:
		return nil, fmt.Errorf("llm: unknown provider %q", cfg.LLM.Provider)
	}
}
