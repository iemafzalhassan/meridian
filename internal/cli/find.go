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
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewFindCmd(deps *Dependencies) *cobra.Command {
	var skills string
	var level string
	cmd := &cobra.Command{
		Use:   "find",
		Short: "Find projects matching your skills",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			deps.Logger.Info("find requested", zap.String("skills", skills))
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
			return agent.RunFind(ctx, deps.Logger, embedder, store, llm, skills, level)
		},
	}
	cmd.Flags().StringVar(&skills, "skills", "", "Comma-separated skills, e.g. \"Go,gRPC,Kubernetes\"")
	cmd.Flags().StringVar(&level, "level", "intermediate", "Experience level: beginner, intermediate, or advanced")
	_ = cmd.MarkFlagRequired("skills")
	return cmd
}
