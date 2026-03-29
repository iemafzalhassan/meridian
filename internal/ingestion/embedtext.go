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
	"strings"

	"github.com/iemafzalhassan/meridian/internal/vectorstore"
)

// BuildEmbedText matches the Meridian concatenation contract for embeddings.
func BuildEmbedText(p *vectorstore.ProjectDoc) {
	if p == nil {
		return
	}
	p.EmbedText = strings.TrimSpace(p.Name) + ". " + strings.TrimSpace(p.Description) +
		". Category: " + strings.TrimSpace(p.Category) + ".\nTechnologies: " +
		strings.Join(p.Languages, ", ") + ". Topics: " + strings.Join(p.Topics, ", ") + "."
}
