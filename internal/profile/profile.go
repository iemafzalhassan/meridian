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

package profile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/iemafzalhassan/meridian/internal/config"
	"gopkg.in/yaml.v3"
)

type UserProfile struct {
	Name       string   `yaml:"name" json:"name"`
	GitHub     string   `yaml:"github" json:"github"`
	Skills     []string `yaml:"skills" json:"skills"`
	Level      string   `yaml:"level" json:"level"`
	Interests  []string `yaml:"interests" json:"interests"`
	LastUpdate string   `yaml:"last_update" json:"last_update"`
}

// DefaultPath returns ~/.meridian/profile.yaml
func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("profile: home dir: %w", err)
	}
	return filepath.Join(home, config.DefaultConfigDirName, "profile.yaml"), nil
}

// Load reads a profile YAML file.
func Load(path string) (*UserProfile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("profile: read %s: %w", path, err)
	}
	var p UserProfile
	if err := yaml.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("profile: parse yaml: %w", err)
	}
	return &p, nil
}

// SkillsString joins skills for embedding/rerank input.
func SkillsString(p *UserProfile) string {
	if p == nil {
		return ""
	}
	return strings.TrimSpace(strings.Join(p.Skills, ", "))
}
