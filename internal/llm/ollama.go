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
	"context"
	"fmt"
	"net/http"

	"github.com/iemafzalhassan/meridian/internal/config"
	"github.com/iemafzalhassan/meridian/internal/retry"
	"github.com/tmc/langchaingo/llms/ollama"
)

type OllamaProvider struct {
	cfg *config.Config
	llm *ollama.LLM
}

func NewOllamaProvider(cfg *config.Config) (*OllamaProvider, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm: missing config")
	}
	httpClient := &http.Client{Timeout: cfg.LLM.Timeout}
	m, err := ollama.New(
		ollama.WithServerURL(cfg.LLM.OllamaURL),
		ollama.WithModel(cfg.LLM.Model),
		ollama.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("llm: ollama client: %w", err)
	}
	return &OllamaProvider{cfg: cfg, llm: m}, nil
}

func (p *OllamaProvider) Complete(ctx context.Context, prompt string) (string, error) {
	var out string
	err := retry.Do(ctx, 1+p.cfg.LLM.RetryCount, p.cfg.LLM.Timeout/5, func(int) error {
		s, err := p.llm.Call(ctx, prompt)
		if err != nil {
			return fmt.Errorf("llm: ollama complete: %w", err)
		}
		out = s
		return nil
	})
	return out, err
}

func (p *OllamaProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	var out []float32
	err := retry.Do(ctx, 1+p.cfg.LLM.RetryCount, p.cfg.LLM.Timeout/5, func(int) error {
		embs, err := p.llm.CreateEmbedding(ctx, []string{text})
		if err != nil {
			return fmt.Errorf("llm: ollama embed: %w", err)
		}
		if len(embs) != 1 || len(embs[0]) == 0 {
			return fmt.Errorf("llm: ollama empty embedding")
		}
		out = embs[0]
		return nil
	})
	return out, err
}
