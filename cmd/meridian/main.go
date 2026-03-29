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

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/iemafzalhassan/meridian/internal/cli"
	"github.com/iemafzalhassan/meridian/internal/config"
	"go.uber.org/zap"
)

var version = "0.1.0-dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		exitWithErr(fmt.Errorf("main: load config: %w", err))
	}

	logger, err := newLogger(cfg.Logging.Level)
	if err != nil {
		exitWithErr(fmt.Errorf("main: setup logger: %w", err))
	}
	defer func() {
		_ = logger.Sync()
	}()

	cmd, err := cli.NewRootCmd(version, &cli.Dependencies{
		Config: cfg,
		Logger: logger,
	})
	if err != nil {
		exitWithErr(fmt.Errorf("main: create root command: %w", err))
	}

	if err := cmd.Execute(); err != nil {
		logger.Error("command failed", zap.Error(err))
		exitWithErr(err)
	}
}

func newLogger(level string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	lvl := strings.ToLower(strings.TrimSpace(level))
	if lvl == "" {
		lvl = "info"
	}
	if err := cfg.Level.UnmarshalText([]byte(lvl)); err != nil {
		return nil, fmt.Errorf("main: invalid log level %q: %w", level, err)
	}
	logger, err := cfg.Build()
	if err != nil {
		return nil, fmt.Errorf("main: build logger: %w", err)
	}
	return logger, nil
}

func exitWithErr(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}
