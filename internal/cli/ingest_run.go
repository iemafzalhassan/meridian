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

package cli

import (
	"context"
	"fmt"

	"github.com/iemafzalhassan/meridian/internal/embeddings"
	"github.com/iemafzalhassan/meridian/internal/ingestion"
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
)

func runIngest(ctx context.Context, deps *Dependencies, limit int, dryRun bool) error {
	embedder, err := embeddings.NewFromConfig(deps.Config)
	if err != nil {
		return fmt.Errorf("cli: embedder: %w", err)
	}
	store := vectorstore.NewQdrantStore(deps.Config)
	defer func() { _ = store.Close() }()
	p := ingestion.NewPipeline(deps.Config, deps.Logger, embedder, store)
	if err := p.Run(ctx, limit, dryRun); err != nil {
		return fmt.Errorf("cli: %w", err)
	}
	return nil
}
