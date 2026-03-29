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

package vectorstore

type ProjectDoc struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Description      string   `json:"description"`
	Category         string   `json:"category"`
	SubCategory      string   `json:"sub_category"`
	RepoURL          string   `json:"repo_url"`
	HomepageURL      string   `json:"homepage_url"`
	Languages        []string `json:"languages"`
	Topics           []string `json:"topics"`
	Stars            int      `json:"stars"`
	GoodFirstIssues  int      `json:"good_first_issues"`
	ContributorCount int      `json:"contributor_count"`
	OpenIssuesCount  int      `json:"open_issues_count"`
	LastCommitDate   string   `json:"last_commit_date"`
	EmbedText        string   `json:"embed_text"`
}

type SearchResult struct {
	Project ProjectDoc `json:"project"`
	Score   float32    `json:"score"`
	Reason  string     `json:"reason,omitempty"`
}
