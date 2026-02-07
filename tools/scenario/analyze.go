// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package scenario

import (
	"fmt"
	"strings"
	"time"
)

// Extract generates a Scenario definition from a session's events.
// It extracts user prompts and computes scoring baselines from actual metrics.
func Extract(sessionID string) (*Scenario, error) {
	se, err := LoadSessionEvents(sessionID)
	if err != nil {
		return nil, err
	}

	userMsgs := se.UserMessages()
	if len(userMsgs) == 0 {
		return nil, fmt.Errorf("session %s has no user messages", sessionID)
	}

	// Build prompts from user messages
	prompts := make([]Prompt, len(userMsgs))
	for i, msg := range userMsgs {
		prompts[i] = Prompt{Text: msg}
	}

	// Compute scoring baselines with headroom
	durationMin := int(se.Duration().Minutes()*1.5) + 1
	if durationMin < 5 {
		durationMin = 5
	}
	turns := int(float64(se.TurnCount()) * 1.3)
	if turns < 10 {
		turns = 10
	}
	azdUps := se.CountToolCallsMatching("powershell", `azd up`)
	maxAzdUps := azdUps + 1
	if maxAzdUps < 3 {
		maxAzdUps = 3
	}
	bicepEdits := se.CountToolCallsMatching("edit", `main\.bicep`)
	maxBicepEdits := bicepEdits + 2
	if maxBicepEdits < 4 {
		maxBicepEdits = 4
	}

	// Generate name from first prompt
	name := slugify(userMsgs[0])

	s := &Scenario{
		Name:        name,
		Description: fmt.Sprintf("Extracted from session %s", sessionID),
		Timeout:     fmt.Sprintf("%dm", durationMin+5),
		Prompts:     prompts,
		Scoring: Scoring{
			MaxDurationMin:   durationMin,
			MaxTurns:         turns,
			MaxAzdUpAttempts: maxAzdUps,
			MaxBicepEdits:    maxBicepEdits,
			MustDelegate:     len(userMsgs) > 1, // multi-prompt = likely standard complexity
			MustInvokeSkills: []string{"avm-bicep-rules"},
			Regressions: []Regression{
				{Name: "ACR auth spiral", Pattern: `ACR.*auth|can't pull|registry.*credential`, MaxOccurrences: 2},
				{Name: "zone redundancy", Pattern: `zone.*redundant|requires.*subnet`, MaxOccurrences: 1},
				{Name: "npm ci without lockfile", Pattern: `npm ci.*lockfile|package-lock.*not found`, MaxOccurrences: 0},
			},
		},
	}

	return s, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, s)
	// collapse multiple dashes
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	if len(s) > 50 {
		s = s[:50]
	}
	return s
}

// Analyze scores a session against a scenario's criteria and returns a Run result.
func Analyze(sessionID string, s *Scenario, gitCommit string) (*Run, error) {
	se, err := LoadSessionEvents(sessionID)
	if err != nil {
		return nil, err
	}

	// Compute metrics
	durationSec := int(se.Duration().Seconds())
	turns := se.TurnCount()
	azdUps := se.CountToolCallsMatching("powershell", `azd up`)
	bicepEdits := se.CountToolCallsMatching("edit", `main\.bicep`)
	delegated := se.HasDelegation()
	skillsInvoked := se.SkillsInvoked()

	// Check skills
	skillResults := make(map[string]bool)
	for _, required := range s.Scoring.MustInvokeSkills {
		found := false
		for _, invoked := range skillsInvoked {
			if invoked == required {
				found = true
				break
			}
		}
		skillResults[required] = found
	}

	// Check regressions
	regResults := make(map[string]RegResult)
	for _, reg := range s.Scoring.Regressions {
		matches := se.CountRegressionMatches(reg.Pattern)
		regResults[reg.Name] = RegResult{
			Occurrences: matches,
			MaxAllowed:  reg.MaxOccurrences,
			Passed:      matches <= reg.MaxOccurrences,
		}
	}

	// Compute pass/fail
	passed := true
	if s.Scoring.MaxDurationMin > 0 && durationSec > s.Scoring.MaxDurationMin*60 {
		passed = false
	}
	if s.Scoring.MaxTurns > 0 && turns > s.Scoring.MaxTurns {
		passed = false
	}
	if s.Scoring.MaxAzdUpAttempts > 0 && azdUps > s.Scoring.MaxAzdUpAttempts {
		passed = false
	}
	if s.Scoring.MaxBicepEdits > 0 && bicepEdits > s.Scoring.MaxBicepEdits {
		passed = false
	}
	if s.Scoring.MustDelegate && !delegated {
		passed = false
	}
	for _, invoked := range skillResults {
		if !invoked {
			passed = false
		}
	}
	for _, reg := range regResults {
		if !reg.Passed {
			passed = false
		}
	}

	// Compute composite score (0.0-1.0)
	score := computeScore(s, durationSec, turns, azdUps, bicepEdits, delegated, skillResults, regResults)

	startedAt := time.Time{}
	if len(se.Events) > 0 {
		startedAt = se.Events[0].Timestamp
	}

	return &Run{
		Scenario:      s.Name,
		SessionID:     sessionID,
		GitCommit:     gitCommit,
		StartedAt:     startedAt,
		DurationSec:   durationSec,
		TotalTurns:    turns,
		AzdUpAttempts: azdUps,
		BicepEdits:    bicepEdits,
		Delegated:     delegated,
		Deployed:      azdUps > 0, // rough proxy: if azd up was called, deployment was attempted
		Score:         score,
		Passed:        passed,
		Skills:        skillResults,
		Regressions:   regResults,
	}, nil
}

func computeScore(s *Scenario, durationSec, turns, azdUps, bicepEdits int,
	delegated bool, skills map[string]bool, regs map[string]RegResult) float64 {

	total := 0.0
	maxPoints := 0.0

	// Continuous scoring: full points at/below limit, proportional reduction above.
	// Uses limit/actual so even 10x over still gets some credit (10%).
	scoreMetric := func(actual, limit, weight float64) float64 {
		maxPoints += weight
		if limit <= 0 {
			total += weight
			return weight
		}
		if actual <= limit {
			total += weight
			return weight
		}
		// Proportional: score = weight * (limit / actual)
		// At 2x over → 50%, at 3x → 33%, at 10x → 10%
		pts := weight * limit / actual
		total += pts
		return pts
	}

	scoreMetric(float64(durationSec), float64(s.Scoring.MaxDurationMin*60), 25)
	scoreMetric(float64(turns), float64(s.Scoring.MaxTurns), 20)
	scoreMetric(float64(azdUps), float64(s.Scoring.MaxAzdUpAttempts), 20)
	scoreMetric(float64(bicepEdits), float64(s.Scoring.MaxBicepEdits), 10)

	// Delegation (10 points)
	if s.Scoring.MustDelegate {
		maxPoints += 10
		if delegated {
			total += 10
		}
	}

	// Skills (5 points per required skill)
	for _, invoked := range skills {
		maxPoints += 5
		if invoked {
			total += 5
		}
	}

	// Regressions (5 points per check)
	for _, reg := range regs {
		maxPoints += 5
		if reg.Passed {
			total += 5
		}
	}

	if maxPoints == 0 {
		return 1.0
	}
	return total / maxPoints
}

// FormatReport generates a markdown report for a run.
func FormatReport(r *Run, s *Scenario) string {
	var b strings.Builder

	status := "✅ PASSED"
	if !r.Passed {
		status = "❌ FAILED"
	}

	fmt.Fprintf(&b, "# Scenario Report: %s\n\n", s.Name)
	fmt.Fprintf(&b, "**Status:** %s | **Score:** %.0f%%\n", status, r.Score*100)
	fmt.Fprintf(&b, "**Session:** %s\n", r.SessionID)
	if r.GitCommit != "" {
		fmt.Fprintf(&b, "**Commit:** %s\n", r.GitCommit)
	}
	fmt.Fprintf(&b, "**Date:** %s\n\n", r.StartedAt.Format("2006-01-02 15:04"))

	fmt.Fprintf(&b, "## Metrics\n\n")
	fmt.Fprintf(&b, "| Metric | Value | Limit | Status |\n")
	fmt.Fprintf(&b, "|--------|-------|-------|--------|\n")

	fmt.Fprintf(&b, "| Duration | %ds | %ds | %s |\n",
		r.DurationSec, s.Scoring.MaxDurationMin*60, passFail(r.DurationSec <= s.Scoring.MaxDurationMin*60))
	fmt.Fprintf(&b, "| Turns | %d | %d | %s |\n",
		r.TotalTurns, s.Scoring.MaxTurns, passFail(r.TotalTurns <= s.Scoring.MaxTurns))
	fmt.Fprintf(&b, "| azd up attempts | %d | %d | %s |\n",
		r.AzdUpAttempts, s.Scoring.MaxAzdUpAttempts, passFail(r.AzdUpAttempts <= s.Scoring.MaxAzdUpAttempts))
	fmt.Fprintf(&b, "| Bicep edits | %d | %d | %s |\n",
		r.BicepEdits, s.Scoring.MaxBicepEdits, passFail(r.BicepEdits <= s.Scoring.MaxBicepEdits))
	if s.Scoring.MustDelegate {
		fmt.Fprintf(&b, "| Delegated | %v | required | %s |\n",
			r.Delegated, passFail(r.Delegated))
	}

	if len(r.Skills) > 0 {
		fmt.Fprintf(&b, "\n## Skills\n\n")
		fmt.Fprintf(&b, "| Skill | Invoked |\n")
		fmt.Fprintf(&b, "|-------|---------|\n")
		for skill, invoked := range r.Skills {
			fmt.Fprintf(&b, "| %s | %s |\n", skill, passFail(invoked))
		}
	}

	if len(r.Regressions) > 0 {
		fmt.Fprintf(&b, "\n## Regressions\n\n")
		fmt.Fprintf(&b, "| Check | Occurrences | Max | Status |\n")
		fmt.Fprintf(&b, "|-------|-------------|-----|--------|\n")
		for name, reg := range r.Regressions {
			fmt.Fprintf(&b, "| %s | %d | %d | %s |\n",
				name, reg.Occurrences, reg.MaxAllowed, passFail(reg.Passed))
		}
	}

	return b.String()
}

func passFail(ok bool) string {
	if ok {
		return "✅"
	}
	return "❌"
}
