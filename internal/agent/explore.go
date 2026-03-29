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
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss/table"
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
)

// RunExplore lists top projects in a category (substring match on CNCF category path).
func RunExplore(ctx context.Context, store *vectorstore.QdrantStore, category string, limit int) error {
	if limit <= 0 {
		limit = 20
	}
	results, err := store.ScrollCategory(ctx, category, limit)
	if err != nil {
		return fmt.Errorf("agent: explore: %w", err)
	}
	if len(results) == 0 {
		fmt.Fprintln(os.Stderr, "No matches. Ingest data with: meridian ingest --limit 100")
		return nil
	}
	t := table.New().Headers("Project", "Stars", "GFI", "Category", "Repo")
	for _, r := range results {
		p := r.Project
		t.Row(p.Name, fmt.Sprintf("%d", p.Stars), fmt.Sprintf("%d", p.GoodFirstIssues),
			truncate(p.Category+" / "+p.SubCategory, 40), p.RepoURL)
	}
	fmt.Println(t.String())
	return nil
}
