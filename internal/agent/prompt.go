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

import "fmt"

const (
	RerankSystemPrompt = `You are Meridian, an open source contribution guide.
Given a developer's skills and a list of candidate projects,
rank the top 5 by fit and explain WHY each project matches.
Be specific - mention exact skills that transfer.
Format: JSON array of {name, reason, good_first_issues, repo_url, score}.`
)

// RerankUser builds the user message for reranking.
func RerankUser(skills, level, candidatesJSON string) string {
	return fmt.Sprintf(`Developer skills: %s
Experience level: %s
Candidate projects: %s`, skills, level, candidatesJSON)
}
