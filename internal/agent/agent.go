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

package agent

import (
	"context"
	"fmt"
)

type LLM interface {
	Complete(ctx context.Context, prompt string) (string, error)
}

type Searcher interface {
	Search(ctx context.Context, embedding []float32, limit int, category string) ([]byte, error)
}

type Agent struct {
	llm LLM
}

func New(llm LLM) *Agent {
	return &Agent{llm: llm}
}

func (a *Agent) Recommend(ctx context.Context, prompt string) (string, error) {
	if a.llm == nil {
		return "", fmt.Errorf("agent: missing llm dependency")
	}
	resp, err := a.llm.Complete(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("agent: complete prompt: %w", err)
	}
	return resp, nil
}
