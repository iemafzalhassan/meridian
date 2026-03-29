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

package embeddings

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/iemafzalhassan/meridian/internal/config"
	"github.com/iemafzalhassan/meridian/internal/retry"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenRouterEmbedder struct {
	cfg *config.Config
	llm *openai.LLM
}

func NewOpenRouterEmbedder(cfg *config.Config) (*OpenRouterEmbedder, error) {
	if cfg == nil {
		return nil, fmt.Errorf("embeddings: missing config")
	}
	key := strings.TrimSpace(cfg.LLM.OpenRouterKey)
	if key == "" {
		return nil, fmt.Errorf("embeddings: openrouter_key / MERIDIAN_OPENROUTER_KEY is required")
	}
	httpClient := &http.Client{Timeout: cfg.Embeddings.Timeout}
	m, err := openai.New(
		openai.WithToken(key),
		openai.WithModel(cfg.Embeddings.Model),
		openai.WithEmbeddingModel(cfg.Embeddings.Model),
		openai.WithBaseURL("https://openrouter.ai/api/v1"),
		openai.WithHTTPClient(httpClient),
	)
	if err != nil {
		return nil, fmt.Errorf("embeddings: openrouter client: %w", err)
	}
	return &OpenRouterEmbedder{cfg: cfg, llm: m}, nil
}

func (e *OpenRouterEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	var out []float32
	err := retry.Do(ctx, 1+e.cfg.Embeddings.RetryCount, e.cfg.Embeddings.Timeout/5, func(int) error {
		embs, err := e.llm.CreateEmbedding(ctx, []string{text})
		if err != nil {
			return fmt.Errorf("embeddings: openrouter embedding: %w", err)
		}
		if len(embs) != 1 || len(embs[0]) == 0 {
			return fmt.Errorf("embeddings: openrouter empty embedding response")
		}
		out = embs[0]
		return nil
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}
