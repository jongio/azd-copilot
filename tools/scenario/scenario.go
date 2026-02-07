// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

// Package scenario provides automated testing for azd-copilot flows.
// It supports defining scenarios as YAML, replaying them via azd copilot,
// analyzing session logs, and tracking results over time in SQLite.
package scenario

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Scenario defines a repeatable test scenario for azd-copilot.
type Scenario struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Timeout     string   `yaml:"timeout,omitempty"` // e.g. "30m"
	Prompts     []Prompt `yaml:"prompts"`
	Scoring     Scoring  `yaml:"scoring"`
}

// Prompt is a single user message injected into the copilot session.
type Prompt struct {
	Text            string          `yaml:"text"`
	SuccessCriteria SuccessCriteria `yaml:"success_criteria,omitempty"`
}

// SuccessCriteria defines what must be true after a prompt completes.
type SuccessCriteria struct {
	FilesExist       []string `yaml:"files_exist,omitempty"`
	Deployed         bool     `yaml:"deployed,omitempty"`
	EndpointResponds bool     `yaml:"endpoint_responds,omitempty"`
}

// Scoring defines how a scenario run is evaluated.
type Scoring struct {
	MaxDurationMin   int          `yaml:"max_duration_minutes,omitempty"`
	MaxTurns         int          `yaml:"max_turns,omitempty"`
	MaxAzdUpAttempts int          `yaml:"max_azd_up_attempts,omitempty"`
	MaxBicepEdits    int          `yaml:"max_bicep_edits,omitempty"`
	MustDelegate     bool         `yaml:"must_delegate,omitempty"`
	MustInvokeSkills []string     `yaml:"must_invoke_skills,omitempty"`
	Regressions      []Regression `yaml:"regressions,omitempty"`
}

// Regression is a named pattern to watch for in assistant messages.
type Regression struct {
	Name           string `yaml:"name"`
	Pattern        string `yaml:"pattern"`
	MaxOccurrences int    `yaml:"max_occurrences"`
}

// LoadScenario reads a scenario YAML file.
func LoadScenario(path string) (*Scenario, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario: %w", err)
	}
	var s Scenario
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse scenario %s: %w", path, err)
	}
	return &s, nil
}

// SaveScenario writes a scenario to a YAML file.
func SaveScenario(s *Scenario, path string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshal scenario: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// --- Session Event Parsing ---

// Event represents a single event from events.jsonl.
type Event struct {
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
	ID        string          `json:"id"`
	Timestamp time.Time       `json:"timestamp"`
	ParentID  *string         `json:"parentId"`
}

// UserMessageData is the data payload for user.message events.
type UserMessageData struct {
	Content string `json:"content"`
}

// AssistantMessageData is the data payload for assistant.message events.
type AssistantMessageData struct {
	Content string `json:"content"`
}

// ToolExecutionData is the data payload for tool.execution_start events.
type ToolExecutionData struct {
	ToolName  string          `json:"toolName"`
	Arguments json.RawMessage `json:"arguments"`
}

// SkillInvokedData is the data payload for skill.invoked events.
type SkillInvokedData struct {
	Name string `json:"name"`
}

// SessionEvents holds parsed events from a session log.
type SessionEvents struct {
	Events []Event
}

// LoadSessionEvents reads and parses events.jsonl for a session.
func LoadSessionEvents(sessionID string) (*SessionEvents, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get home dir: %w", err)
	}

	eventsPath := filepath.Join(home, ".copilot", "session-state", sessionID, "events.jsonl")
	data, err := os.ReadFile(eventsPath)
	if err != nil {
		return nil, fmt.Errorf("read events.jsonl for session %s: %w", sessionID, err)
	}

	var events []Event
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var e Event
		if err := json.Unmarshal([]byte(line), &e); err != nil {
			continue // skip malformed lines
		}
		events = append(events, e)
	}

	return &SessionEvents{Events: events}, nil
}

// UserMessages returns the content of all user.message events.
func (se *SessionEvents) UserMessages() []string {
	var msgs []string
	for _, e := range se.Events {
		if e.Type == "user.message" {
			var d UserMessageData
			if json.Unmarshal(e.Data, &d) == nil {
				msgs = append(msgs, d.Content)
			}
		}
	}
	return msgs
}

// AssistantMessages returns the content of all assistant.message events.
func (se *SessionEvents) AssistantMessages() []string {
	var msgs []string
	for _, e := range se.Events {
		if e.Type == "assistant.message" {
			var d AssistantMessageData
			if json.Unmarshal(e.Data, &d) == nil {
				msgs = append(msgs, d.Content)
			}
		}
	}
	return msgs
}

// TurnCount returns the number of assistant turns.
func (se *SessionEvents) TurnCount() int {
	n := 0
	for _, e := range se.Events {
		if e.Type == "assistant.turn_start" {
			n++
		}
	}
	return n
}

// ToolCalls returns tool names for all tool.execution_start events.
func (se *SessionEvents) ToolCalls() []ToolExecutionData {
	var calls []ToolExecutionData
	for _, e := range se.Events {
		if e.Type == "tool.execution_start" {
			var d ToolExecutionData
			if json.Unmarshal(e.Data, &d) == nil {
				calls = append(calls, d)
			}
		}
	}
	return calls
}

// SkillsInvoked returns the names of all skills invoked.
func (se *SessionEvents) SkillsInvoked() []string {
	var skills []string
	for _, e := range se.Events {
		if e.Type == "skill.invoked" {
			var d SkillInvokedData
			if json.Unmarshal(e.Data, &d) == nil {
				skills = append(skills, d.Name)
			}
		}
	}
	return skills
}

// Duration returns the time between first and last event.
func (se *SessionEvents) Duration() time.Duration {
	if len(se.Events) < 2 {
		return 0
	}
	return se.Events[len(se.Events)-1].Timestamp.Sub(se.Events[0].Timestamp)
}

// CountToolCallsMatching counts tool.execution_start events where the command matches a pattern.
func (se *SessionEvents) CountToolCallsMatching(toolName string, argPattern string) int {
	n := 0
	for _, tc := range se.ToolCalls() {
		if tc.ToolName != toolName {
			continue
		}
		if argPattern == "" {
			n++
			continue
		}
		if matched, _ := regexp.MatchString(argPattern, string(tc.Arguments)); matched {
			n++
		}
	}
	return n
}

// HasDelegation checks if any task() tool calls were made to agent_type agents.
func (se *SessionEvents) HasDelegation() bool {
	for _, tc := range se.ToolCalls() {
		if tc.ToolName == "task" {
			return true
		}
	}
	return false
}

// CountRegressionMatches counts how many assistant messages match a regression pattern.
func (se *SessionEvents) CountRegressionMatches(pattern string) int {
	re, err := regexp.Compile("(?i)" + pattern)
	if err != nil {
		return 0
	}
	n := 0
	for _, msg := range se.AssistantMessages() {
		if re.MatchString(msg) {
			n++
		}
	}
	return n
}
