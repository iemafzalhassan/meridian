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

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss/table"
	"github.com/iemafzalhassan/meridian/internal/embeddings"
	llmprovider "github.com/iemafzalhassan/meridian/internal/llm"
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
	"go.uber.org/zap"
)

// RankedProject is the structured JSON the rerank prompt asks for.
type RankedProject struct {
	Name            string  `json:"name"`
	Reason          string  `json:"reason"`
	GoodFirstIssues int     `json:"good_first_issues"`
	RepoURL         string  `json:"repo_url"`
	Score           float64 `json:"score"`
}

type candidateDTO struct {
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Category         string   `json:"category"`
	SubCategory      string   `json:"sub_category"`
	RepoURL          string   `json:"repo_url"`
	Stars            int      `json:"stars"`
	GoodFirstIssues  int      `json:"good_first_issues"`
	Languages        []string `json:"languages"`
	Topics           []string `json:"topics"`
	ContributorCount int      `json:"contributor_count"`
}

// RunFind embeds skills, searches Qdrant, optionally reranks with the LLM, and prints a table.
func RunFind(ctx context.Context, log *zap.Logger, embedder embeddings.Embedder, store *vectorstore.QdrantStore, llm llmprovider.Provider, skills, level string) error {
	skills = strings.TrimSpace(skills)
	if skills == "" {
		return fmt.Errorf("agent: skills are required")
	}
	level = strings.TrimSpace(level)
	if level == "" {
		level = "intermediate"
	}
	vec, err := embedder.Embed(ctx, skills)
	if err != nil {
		return fmt.Errorf("agent: embed skills: %w", err)
	}
	results, err := store.SearchSimilar(ctx, vec, 20, "")
	if err != nil {
		return fmt.Errorf("agent: vector search: %w", err)
	}
	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "No projects in Qdrant yet. Run: meridian ingest --limit 50")
		return nil
	}
	cands := make([]candidateDTO, 0, len(results))
	for _, r := range results {
		p := r.Project
		cands = append(cands, candidateDTO{
			Name:             p.Name,
			Description:      truncate(p.Description, 240),
			Category:         p.Category,
			SubCategory:      p.SubCategory,
			RepoURL:          p.RepoURL,
			Stars:            p.Stars,
			GoodFirstIssues:  p.GoodFirstIssues,
			Languages:        p.Languages,
			Topics:           p.Topics,
			ContributorCount: p.ContributorCount,
		})
	}
	cjsonb, err := json.Marshal(cands)
	if err != nil {
		return fmt.Errorf("agent: marshal candidates: %w", err)
	}
	cjson := string(cjsonb)
	fullPrompt := RerankSystemPrompt + "\n\n" + RerankUser(skills, level, cjson)
	var ranked []RankedProject
	if llm != nil {
		raw, err := llm.Complete(ctx, fullPrompt)
		if err != nil {
			log.Warn("agent: rerank llm failed; showing vector results only", zap.Error(err))
		} else {
			clean := vectorstore.StripJSONFences(raw)
			if uerr := json.Unmarshal([]byte(clean), &ranked); uerr != nil {
				log.Warn("agent: parse rerank json failed; showing vector results only", zap.Error(uerr))
			}
		}
	}
	if len(ranked) == 0 {
		return printVectorFallback(results)
	}
	if len(ranked) > 5 {
		ranked = ranked[:5]
	}
	return printRankedTable(ranked)
}

func printVectorFallback(results []vectorstore.SearchResult) error {
	t := table.New().Headers("Project", "Fit (sim)", "GFI", "Repo")
	n := 0
	for _, r := range results {
		if n >= 20 {
			break
		}
		t.Row(r.Project.Name,
			fmt.Sprintf("%.3f", r.Score),
			fmt.Sprintf("%d", r.Project.GoodFirstIssues),
			r.Project.RepoURL,
		)
		n++
	}
	fmt.Println(t.String())
	return nil
}

func printRankedTable(rank []RankedProject) error {
	t := table.New().Headers("Project", "Score", "GFI", "Why", "Repo")
	for _, r := range rank {
		t.Row(r.Name,
			fmt.Sprintf("%.1f", r.Score),
			fmt.Sprintf("%d", r.GoodFirstIssues),
			truncate(r.Reason, 80),
			r.RepoURL,
		)
	}
	fmt.Println(t.String())
	return nil
}

func truncate(s string, max int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
