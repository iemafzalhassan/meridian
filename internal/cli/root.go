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

package cli

import (
	"fmt"

	"github.com/iemafzalhassan/meridian/internal/config"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type Dependencies struct {
	Config *config.Config
	Logger *zap.Logger
}

func NewRootCmd(version string, deps *Dependencies) (*cobra.Command, error) {
	if deps == nil {
		return nil, fmt.Errorf("cli: dependencies are required")
	}
	if deps.Config == nil {
		return nil, fmt.Errorf("cli: config dependency is required")
	}
	if deps.Logger == nil {
		return nil, fmt.Errorf("cli: logger dependency is required")
	}

	rootCmd := &cobra.Command{
		Use:   "meridian",
		Short: "Find your place in open source.",
		Long: `Meridian is an AI-powered open source contribution intelligence tool.
It helps developers discover projects where their skills fit best.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
	}

	rootCmd.AddCommand(
		NewInitCmd(deps),
		NewFindCmd(deps),
		NewExploreCmd(deps),
		NewRecommendCmd(deps),
		newIngestCmd(deps),
	)

	return rootCmd, nil
}

func newIngestCmd(deps *Dependencies) *cobra.Command {
	var limit int
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Ingest and enrich OSS project data into Qdrant",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = args
			deps.Logger.Info("ingest requested", zap.Int("limit", limit), zap.Bool("dry_run", dryRun))
			return runIngest(cmd.Context(), deps, limit, dryRun)
		},
	}
	cmd.Flags().IntVar(&limit, "limit", 100, "Limit number of projects for dev ingestion")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Validate pipeline without writing to Qdrant")

	return cmd
}
