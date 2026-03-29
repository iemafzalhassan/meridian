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
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v67/github"
	"github.com/iemafzalhassan/meridian/internal/config"
	"github.com/iemafzalhassan/meridian/internal/retry"
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
)

var githubRepoRE = regexp.MustCompile(`(?i)github\.com/([^/]+)/([^/.]+)`)

// GitHubEnricher loads repository metadata into ProjectDoc fields.
type GitHubEnricher struct {
	cfg    *config.Config
	client *github.Client
}

func NewGitHubEnricher(cfg *config.Config) *GitHubEnricher {
	httpClient := &http.Client{Timeout: cfg.GitHub.Timeout}
	gh := github.NewClient(httpClient)
	if strings.TrimSpace(cfg.GitHub.Token) != "" {
		gh = gh.WithAuthToken(cfg.GitHub.Token)
	}
	return &GitHubEnricher{cfg: cfg, client: gh}
}

// Enrich mutates doc with GitHub API data (best effort for each field).
func (g *GitHubEnricher) Enrich(ctx context.Context, doc *vectorstore.ProjectDoc) error {
	owner, repo, ok := parseGitHubRepo(doc.RepoURL)
	if !ok {
		return fmt.Errorf("ingestion: not a GitHub repo url: %s", doc.RepoURL)
	}
	var r *github.Repository
	err := retry.Do(ctx, 1+g.cfg.GitHub.RetryCount, g.cfg.GitHub.Timeout/5, func(int) error {
		var err error
		r, _, err = g.client.Repositories.Get(ctx, owner, repo)
		if err != nil {
			if rateErr := sleepIfRateLimited(err); rateErr != nil {
				return rateErr
			}
			return fmt.Errorf("ingestion: get repo: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	if r != nil {
		doc.Stars = r.GetStargazersCount()
		doc.OpenIssuesCount = r.GetOpenIssuesCount()
		if t := r.GetPushedAt(); !t.IsZero() {
			doc.LastCommitDate = t.Format(time.RFC3339)
		}
		if topics := r.Topics; len(topics) > 0 {
			doc.Topics = topics
		}
	}

	// Languages
	langs := make(map[string]int)
	_ = retry.Do(ctx, 1+g.cfg.GitHub.RetryCount, g.cfg.GitHub.Timeout/5, func(int) error {
		l, _, err := g.client.Repositories.ListLanguages(ctx, owner, repo)
		if err != nil {
			if rateErr := sleepIfRateLimited(err); rateErr != nil {
				return rateErr
			}
			return fmt.Errorf("ingestion: list languages: %w", err)
		}
		langs = l
		return nil
	})
	doc.Languages = languageKeysSortedBySize(langs)

	// Good first issues (search API)
	_ = retry.Do(ctx, 1+g.cfg.GitHub.RetryCount, g.cfg.GitHub.Timeout/5, func(int) error {
		q := fmt.Sprintf(`repo:%s/%s is:open label:"good first issue"`, owner, repo)
		sr, _, err := g.client.Search.Issues(ctx, q, &github.SearchOptions{ListOptions: github.ListOptions{PerPage: 1}})
		if err != nil {
			if rateErr := sleepIfRateLimited(err); rateErr != nil {
				return rateErr
			}
			return fmt.Errorf("ingestion: search good first issues: %w", err)
		}
		doc.GoodFirstIssues = sr.GetTotal()
		return nil
	})

	// Contributor count (capped list)
	doc.ContributorCount = g.contributorCount(ctx, owner, repo)

	return nil
}

func (g *GitHubEnricher) contributorCount(ctx context.Context, owner, repo string) int {
	total := 0
	opt := &github.ListContributorsOptions{ListOptions: github.ListOptions{PerPage: 100}}
	for page := 0; page < 10; page++ {
		var contribs []*github.Contributor
		err := retry.Do(ctx, 1+g.cfg.GitHub.RetryCount, g.cfg.GitHub.Timeout/5, func(int) error {
			var err error
			contribs, _, err = g.client.Repositories.ListContributors(ctx, owner, repo, opt)
			if err != nil {
				if rateErr := sleepIfRateLimited(err); rateErr != nil {
					return rateErr
				}
				return fmt.Errorf("ingestion: list contributors: %w", err)
			}
			return nil
		})
		if err != nil || len(contribs) == 0 {
			break
		}
		total += len(contribs)
		if len(contribs) < 100 {
			break
		}
		opt.Page++
	}
	return total
}

func parseGitHubRepo(repoURL string) (owner, name string, ok bool) {
	m := githubRepoRE.FindStringSubmatch(strings.TrimSpace(repoURL))
	if len(m) != 3 {
		return "", "", false
	}
	return m[1], strings.TrimSuffix(m[2], ".git"), true
}

func languageKeysSortedBySize(m map[string]int) []string {
	type kv struct {
		k string
		v int
	}
	var pairs []kv
	for k, v := range m {
		pairs = append(pairs, kv{k: k, v: v})
	}
	// simple sort by v desc
	for i := 1; i < len(pairs); i++ {
		for j := i; j > 0 && pairs[j].v > pairs[j-1].v; j-- {
			pairs[j], pairs[j-1] = pairs[j-1], pairs[j]
		}
	}
	out := make([]string, 0, len(pairs))
	for _, p := range pairs {
		out = append(out, p.k)
	}
	return out
}

func sleepIfRateLimited(err error) error {
	if _, ok := err.(*github.RateLimitError); ok {
		time.Sleep(2 * time.Second)
		return err
	}
	if ae, ok := err.(*github.AbuseRateLimitError); ok {
		if ae.GetRetryAfter() > 0 {
			time.Sleep(ae.GetRetryAfter())
		} else {
			time.Sleep(5 * time.Second)
		}
		return err
	}
	return nil
}
