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
	"time"

	"github.com/iemafzalhassan/meridian/internal/agent"
	"github.com/iemafzalhassan/meridian/internal/embeddings"
	llmprovider "github.com/iemafzalhassan/meridian/internal/llm"
	"github.com/iemafzalhassan/meridian/internal/profile"
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
	"github.com/spf13/cobra"
)

func NewRecommendCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "recommend",
		Short: "Recommend projects based on your saved profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			deps.Logger.Info("recommend requested")
			path, err := profile.DefaultPath()
			if err != nil {
				return fmt.Errorf("cli: profile path: %w", err)
			}
			pro, err := profile.Load(path)
			if err != nil {
				return fmt.Errorf("cli: load profile from %s (run meridian init): %w", path, err)
			}
			skills := profile.SkillsString(pro)
			if skills == "" {
				return fmt.Errorf("cli: profile has no skills; run meridian init")
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 3*time.Minute)
			defer cancel()
			embedder, err := embeddings.NewFromConfig(deps.Config)
			if err != nil {
				return fmt.Errorf("cli: embedder: %w", err)
			}
			llm, err := llmprovider.NewFromConfig(deps.Config)
			if err != nil {
				return fmt.Errorf("cli: llm: %w", err)
			}
			store := vectorstore.NewQdrantStore(deps.Config)
			defer func() { _ = store.Close() }()
			level := pro.Level
			if level == "" {
				level = "intermediate"
			}
			if len(pro.Interests) > 0 {
				skills = skills + "; interests: " + profile.SkillsString(&profile.UserProfile{Skills: pro.Interests})
			}
			return agent.RunFind(ctx, deps.Logger, embedder, store, llm, skills, level)
		},
	}
}
