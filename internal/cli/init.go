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

	"github.com/charmbracelet/glamour"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

func NewInitCmd(deps *Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize your Meridian profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			_ = cmd
			_ = args
			md := "# Meridian profile setup\n\nInteractive prompts and saving to `~/.meridian/profile.yaml` are coming in a later milestone.\n\n**For now:** copy `.meridian.yaml.example` to `~/.meridian/config.yaml`, run `meridian ingest`, then `meridian find`.\n"
			out, gerr := glamour.Render(md, "")
			if gerr != nil {
				return fmt.Errorf("cli: render welcome: %w", gerr)
			}
			model := initModel{message: out}
			p := tea.NewProgram(model)
			if _, err := p.Run(); err != nil {
				return fmt.Errorf("cli: run init tui: %w", err)
			}
			deps.Logger.Info("init command finished")
			return nil
		},
	}
}

type initModel struct {
	message string
}

func (m initModel) Init() tea.Cmd {
	return nil
}

func (m initModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" || msg.String() == "enter" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m initModel) View() string {
	return m.message + "\nPress Enter to continue.\n"
}
