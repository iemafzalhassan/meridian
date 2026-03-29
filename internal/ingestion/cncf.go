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
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/iemafzalhassan/meridian/internal/retry"
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
	"gopkg.in/yaml.v3"
)

const LandscapeURL = "https://raw.githubusercontent.com/cncf/landscape/master/landscape.yml"

// FetchLandscapeYAML downloads landscape.yml bytes (with retry).
func FetchLandscapeYAML(ctx context.Context, timeout time.Duration, retries int) ([]byte, error) {
	client := resty.New()
	client.SetTimeout(timeout)
	var body []byte
	err := retry.Do(ctx, 1+retries, timeout/5, func(int) error {
		resp, err := client.R().
			SetContext(ctx).
			Get(LandscapeURL)
		if err != nil {
			return fmt.Errorf("ingestion: http get landscape: %w", err)
		}
		if resp.IsError() {
			return fmt.Errorf("ingestion: landscape http %d", resp.StatusCode())
		}
		body = resp.Body()
		return nil
	})
	if err != nil {
		return nil, err
	}
	return body, nil
}

// ParseLandscape extracts project entries with GitHub repos from raw landscape YAML.
func ParseLandscape(raw []byte) ([]vectorstore.ProjectDoc, error) {
	var root map[string]any
	if err := yaml.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("ingestion: unmarshal landscape root: %w", err)
	}
	landscape, ok := root["landscape"].([]any)
	if !ok {
		return nil, fmt.Errorf("ingestion: landscape key missing or wrong type")
	}
	var out []vectorstore.ProjectDoc
	for _, top := range landscape {
		entry, ok := top.(map[string]any)
		if !ok {
			continue
		}
		catNode, ok := entry["category"].(map[string]any)
		if !ok {
			continue
		}
		out = append(out, walkCategory(catNode)...)
	}
	return out, nil
}

func walkCategory(cat map[string]any) []vectorstore.ProjectDoc {
	categoryName := stringField(cat["name"])
	var docs []vectorstore.ProjectDoc
	subs, ok := cat["subcategories"].([]any)
	if !ok {
		return docs
	}
	for _, s := range subs {
		sm, ok := s.(map[string]any)
		if !ok {
			continue
		}
		sub, ok := sm["subcategory"].(map[string]any)
		if !ok {
			continue
		}
		subName := stringField(sub["name"])
		items, ok := sub["items"].([]any)
		if !ok {
			continue
		}
		for _, it := range items {
			im, ok := it.(map[string]any)
			if !ok {
				continue
			}
			item, ok := im["item"].(map[string]any)
			if !ok {
				continue
			}
			repo := stringField(item["repo_url"])
			if repo == "" || !strings.Contains(strings.ToLower(repo), "github.com") {
				continue
			}
			docs = append(docs, vectorstore.ProjectDoc{
				ID:          repo,
				Name:        stringField(item["name"]),
				Description: stringField(item["description"]),
				Category:    categoryName,
				SubCategory: subName,
				RepoURL:     strings.TrimSuffix(repo, ".git"),
				HomepageURL: stringField(item["homepage_url"]),
			})
		}
	}
	return docs
}

func stringField(v any) string {
	if v == nil {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}
