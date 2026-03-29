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

package ingestion

import (
	"context"
	"fmt"
	"sync"

	"github.com/iemafzalhassan/meridian/internal/config"
	"github.com/iemafzalhassan/meridian/internal/embeddings"
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
	"go.uber.org/zap"
)

const enrichmentWorkers = 6

// Pipeline loads CNCF landscape projects, enriches with GitHub, embeds, and upserts to Qdrant.
type Pipeline struct {
	cfg      *config.Config
	log      *zap.Logger
	enricher *GitHubEnricher
	embedder embeddings.Embedder
	store    *vectorstore.QdrantStore
}

func NewPipeline(cfg *config.Config, log *zap.Logger, embedder embeddings.Embedder, store *vectorstore.QdrantStore) *Pipeline {
	return &Pipeline{
		cfg:      cfg,
		log:      log,
		enricher: NewGitHubEnricher(cfg),
		embedder: embedder,
		store:    store,
	}
}

func (p *Pipeline) Run(ctx context.Context, limit int, dryRun bool) error {
	raw, err := FetchLandscapeYAML(ctx, p.cfg.GitHub.Timeout, p.cfg.GitHub.RetryCount)
	if err != nil {
		return fmt.Errorf("ingestion: %w", err)
	}
	docs, err := ParseLandscape(raw)
	if err != nil {
		return fmt.Errorf("ingestion: %w", err)
	}
	if limit > 0 && len(docs) > limit {
		docs = docs[:limit]
	}
	p.log.Info("ingestion: landscape entries", zap.Int("count", len(docs)))

	jobs := make(chan *vectorstore.ProjectDoc, len(docs))
	var wg sync.WaitGroup
	var mu sync.Mutex
	enriched := make([]vectorstore.ProjectDoc, 0, len(docs))

	worker := func() {
		defer wg.Done()
		for doc := range jobs {
			if doc == nil {
				continue
			}
			if err := p.enricher.Enrich(ctx, doc); err != nil {
				p.log.Debug("ingestion: enrich skipped", zap.String("repo", doc.RepoURL), zap.Error(err))
				continue
			}
			BuildEmbedText(doc)
			mu.Lock()
			enriched = append(enriched, *doc)
			mu.Unlock()
		}
	}
	wg.Add(enrichmentWorkers)
	for i := 0; i < enrichmentWorkers; i++ {
		go worker()
	}
	for i := range docs {
		di := docs[i]
		jobs <- &di
	}
	close(jobs)
	wg.Wait()

	if dryRun {
		p.log.Info("ingestion: dry-run complete", zap.Int("enriched", len(enriched)))
		return nil
	}
	if len(enriched) == 0 {
		return fmt.Errorf("ingestion: no enriched projects; check GitHub token/rate limits")
	}

	dim := uint64(p.cfg.Qdrant.EmbeddingDims)
	if dim == 0 {
		return fmt.Errorf("ingestion: qdrant.embedding_dims must be set (768 for nomic-embed-text, 1536 for text-embedding-3-small)")
	}
	if err := p.store.EnsureCollection(ctx, dim); err != nil {
		return fmt.Errorf("ingestion: %w", err)
	}

	vecs := make([][]float32, 0, len(enriched))
	finalDocs := make([]vectorstore.ProjectDoc, 0, len(enriched))
	for i := range enriched {
		vec, err := p.embedder.Embed(ctx, enriched[i].EmbedText)
		if err != nil {
			p.log.Warn("ingestion: embed failed", zap.String("repo", enriched[i].RepoURL), zap.Error(err))
			continue
		}
		if uint64(len(vec)) != dim {
			return fmt.Errorf("ingestion: embedding width %d does not match qdrant.embedding_dims %d", len(vec), dim)
		}
		vecs = append(vecs, vec)
		finalDocs = append(finalDocs, enriched[i])
	}
	if err := p.store.Upsert(ctx, finalDocs, vecs); err != nil {
		return fmt.Errorf("ingestion: %w", err)
	}
	p.log.Info("ingestion: upsert complete", zap.Int("points", len(finalDocs)))
	return nil
}
