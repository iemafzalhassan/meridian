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
	"time"

	"github.com/iemafzalhassan/meridian/internal/agent"
	"github.com/iemafzalhassan/meridian/internal/vectorstore"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewExploreCmd(deps *Dependencies) *cobra.Command {
	var category string
	var limit int
	cmd := &cobra.Command{
		Use:   "explore",
		Short: "Explore projects by category",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			deps.Logger.Info("explore requested", zap.String("category", category))
			ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
			defer cancel()
			store := vectorstore.NewQdrantStore(deps.Config)
			defer func() { _ = store.Close() }()
			return agent.RunExplore(ctx, store, category, limit)
		},
	}
	cmd.Flags().StringVar(&category, "category", "", "Filter by project category (substring, case-insensitive)")
	cmd.Flags().IntVar(&limit, "limit", 20, "Max number of projects to show")
	_ = cmd.MarkFlagRequired("category")
	return cmd
}
