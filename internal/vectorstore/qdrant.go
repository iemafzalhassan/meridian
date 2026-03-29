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

package vectorstore

import (
	"context"
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/iemafzalhassan/meridian/internal/config"
	"github.com/iemafzalhassan/meridian/internal/qdrantutil"
	"github.com/iemafzalhassan/meridian/internal/retry"
	"github.com/qdrant/go-client/qdrant"
)

// QdrantStore wraps the official Qdrant Go client.
type QdrantStore struct {
	cfg    *config.Config
	client *qdrant.Client
}

func NewQdrantStore(cfg *config.Config) *QdrantStore {
	return &QdrantStore{cfg: cfg}
}

func (s *QdrantStore) connect(ctx context.Context) error {
	if s.client != nil {
		return nil
	}
	host, port, err := qdrantutil.HostPort(s.cfg.Qdrant.URL)
	if err != nil {
		return fmt.Errorf("vectorstore: qdrant address: %w", err)
	}
	var last error
	err = retry.Do(ctx, 1+s.cfg.Qdrant.RetryCount, s.cfg.Qdrant.Timeout/4, func(attempt int) error {
		cli, err := qdrant.NewClient(&qdrant.Config{
			Host: host,
			Port: port,
		})
		if err != nil {
			last = err
			return fmt.Errorf("vectorstore: qdrant dial: %w", err)
		}
		s.client = cli
		last = nil
		return nil
	})
	if err != nil {
		return err
	}
	return last
}

// Close releases the underlying gRPC connection.
func (s *QdrantStore) Close() error {
	if s.client == nil {
		return nil
	}
	err := s.client.Close()
	s.client = nil
	return err
}

// EnsureCollection creates the vector collection when missing.
func (s *QdrantStore) EnsureCollection(ctx context.Context, vectorSize uint64) error {
	if err := s.connect(ctx); err != nil {
		return err
	}
	name := s.cfg.Qdrant.Collection
	if name == "" {
		return fmt.Errorf("vectorstore: collection name is empty")
	}
	var exists bool
	err := retry.Do(ctx, 1+s.cfg.Qdrant.RetryCount, s.cfg.Qdrant.Timeout/4, func(int) error {
		var err error
		exists, err = s.client.CollectionExists(ctx, name)
		return err
	})
	if err != nil {
		return fmt.Errorf("vectorstore: collection exists: %w", err)
	}
	if exists {
		return nil
	}
	err = retry.Do(ctx, 1+s.cfg.Qdrant.RetryCount, s.cfg.Qdrant.Timeout/4, func(int) error {
		return s.client.CreateCollection(ctx, &qdrant.CreateCollection{
			CollectionName: name,
			VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
				Size:     vectorSize,
				Distance: qdrant.Distance_Cosine,
			}),
		})
	})
	if err != nil {
		return fmt.Errorf("vectorstore: create collection: %w", err)
	}
	return nil
}

func pointIDNum(repoURL string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.TrimSpace(strings.TrimSuffix(repoURL, ".git"))))
	return h.Sum64()
}

// Upsert writes or updates points with dense vectors.
func (s *QdrantStore) Upsert(ctx context.Context, docs []ProjectDoc, vectors [][]float32) error {
	if len(docs) != len(vectors) {
		return fmt.Errorf("vectorstore: docs and vectors length mismatch")
	}
	if err := s.connect(ctx); err != nil {
		return err
	}
	wait := true
	points := make([]*qdrant.PointStruct, 0, len(docs))
	for i := range docs {
		pl, err := ProjectDocPayload(docs[i])
		if err != nil {
			return err
		}
		payload, err := qdrant.TryValueMap(pl)
		if err != nil {
			return fmt.Errorf("vectorstore: build payload: %w", err)
		}
		vec := vectors[i]
		points = append(points, &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(pointIDNum(docs[i].RepoURL)),
			Vectors: qdrant.NewVectors(vec...),
			Payload: payload,
		})
	}
	err := retry.Do(ctx, 1+s.cfg.Qdrant.RetryCount, s.cfg.Qdrant.Timeout/4, func(int) error {
		_, err := s.client.Upsert(ctx, &qdrant.UpsertPoints{
			CollectionName: s.cfg.Qdrant.Collection,
			Wait:           &wait,
			Points:         points,
		})
		return err
	})
	if err != nil {
		return fmt.Errorf("vectorstore: upsert: %w", err)
	}
	return nil
}

// SearchSimilar runs a dense nearest-neighbors query (optional category filter uses substring on stored category).
func (s *QdrantStore) SearchSimilar(ctx context.Context, embedding []float32, limit int, categorySubstring string) ([]SearchResult, error) {
	if err := s.connect(ctx); err != nil {
		return nil, err
	}
	if limit <= 0 {
		limit = 10
	}
	fetch := uint64(limit * 4)
	if fetch < uint64(limit) {
		fetch = uint64(limit)
	}
	if fetch > 200 {
		fetch = 200
	}
	req := &qdrant.QueryPoints{
		CollectionName: s.cfg.Qdrant.Collection,
		Query:          qdrant.NewQueryDense(embedding),
		Limit:          qdrant.PtrOf(fetch),
		WithPayload:    qdrant.NewWithPayload(true),
	}
	var hits []*qdrant.ScoredPoint
	err := retry.Do(ctx, 1+s.cfg.Qdrant.RetryCount, s.cfg.Qdrant.Timeout/4, func(int) error {
		var err error
		hits, err = s.client.Query(ctx, req)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("vectorstore: query: %w", err)
	}
	out := make([]SearchResult, 0, limit)
	catNeedle := strings.ToLower(strings.TrimSpace(categorySubstring))
	for _, sp := range hits {
		doc, err := ProjectDocFromPayload(sp.GetPayload())
		if err != nil {
			continue
		}
		if catNeedle != "" {
			cat := strings.ToLower(doc.Category + " " + doc.SubCategory)
			if !strings.Contains(cat, catNeedle) {
				continue
			}
		}
		out = append(out, SearchResult{
			Project: doc,
			Score:   sp.GetScore(),
		})
		if len(out) >= limit {
			break
		}
	}
	return out, nil
}

// ScrollCategory returns projects whose category or subcategory contains needle (case-insensitive), sorted by stars.
func (s *QdrantStore) ScrollCategory(ctx context.Context, needle string, limit int) ([]SearchResult, error) {
	if err := s.connect(ctx); err != nil {
		return nil, err
	}
	needle = strings.ToLower(strings.TrimSpace(needle))
	if needle == "" {
		return nil, fmt.Errorf("vectorstore: category filter is empty")
	}
	const page = uint32(512)
	var offset *qdrant.PointId
	collected := make([]SearchResult, 0, limit)
	for len(collected) < limit*3 && len(collected) < 5000 {
		scrollResp, err := s.client.GetPointsClient().Scroll(ctx, &qdrant.ScrollPoints{
			CollectionName: s.cfg.Qdrant.Collection,
			Limit:          qdrant.PtrOf(page),
			Offset:         offset,
			WithPayload:    qdrant.NewWithPayload(true),
		})
		if err != nil {
			return nil, fmt.Errorf("vectorstore: scroll: %w", err)
		}
		points := scrollResp.GetResult()
		if len(points) == 0 {
			break
		}
		for _, rp := range points {
			doc, err := ProjectDocFromPayload(rp.GetPayload())
			if err != nil {
				continue
			}
			hay := strings.ToLower(doc.Category + " " + doc.SubCategory)
			if strings.Contains(hay, needle) {
				collected = append(collected, SearchResult{Project: doc, Score: 0})
			}
		}
		offset = scrollResp.GetNextPageOffset()
		if offset == nil {
			break
		}
	}
	// Sort by stars descending (simple insertion sort for small n)
	for i := 1; i < len(collected); i++ {
		for j := i; j > 0 && collected[j].Project.Stars > collected[j-1].Project.Stars; j-- {
			collected[j], collected[j-1] = collected[j-1], collected[j]
		}
	}
	if len(collected) > limit {
		collected = collected[:limit]
	}
	return collected, nil
}
