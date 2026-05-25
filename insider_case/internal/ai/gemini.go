package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type GeminiGenerator struct {
	apiKey   string
	model    string
	endpoint string
	client   *http.Client
	fallback CommentaryGenerator
}

func NewGeminiGenerator(apiKey, model string) CommentaryGenerator {
	if strings.TrimSpace(apiKey) == "" {
		return NewDisabledGenerator()
	}
	if strings.TrimSpace(model) == "" {
		model = "gemini-1.5-flash"
	}
	return &GeminiGenerator{
		apiKey:   apiKey,
		model:    model,
		endpoint: "https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		client:   &http.Client{Timeout: 5 * time.Second},
		fallback: NewDegradedGenerator(),
	}
}

func (g *GeminiGenerator) Generate(ctx context.Context, input Input) Output {
	fb := g.fallback.Generate(ctx, input)

	leagueText, err := g.generateText(ctx, g.leaguePrompt(input))
	if err != nil {
		return fb
	}
	nextText, err := g.generateText(ctx, g.nextWeekPrompt(input))
	if err != nil {
		return fb
	}

	return Output{
		League:   strings.TrimSpace(leagueText),
		NextWeek: strings.TrimSpace(nextText),
		Provider: "gemini",
		Status:   "ok",
	}
}

// leaguePrompt builds the title-race analysis prompt.
// Framing is rotated by time so consecutive calls produce different registers
// even with identical data — avoids the "always sounds the same" problem.
func (g *GeminiGenerator) leaguePrompt(input Input) string {
	framings := []string{
		"You are writing a post-round column for a football newsletter.",
		"You are a pundit on a live radio debate about the title race.",
		"You are writing a tactical breakdown for a match-day programme.",
		"You are a journalist filing a 60-word match report summary.",
	}
	framing := framings[time.Now().UnixNano()%int64(len(framings))]

	var b strings.Builder
	b.WriteString(framing + "\n")
	b.WriteString("Cover the title race in 2–3 sentences. Use the actual numbers. Pick a winner and say why — no hedging.\n\n")

	b.WriteString("Table:\n")
	for _, s := range input.Standings {
		b.WriteString(fmt.Sprintf("  %s  %dpts  GD%+d  GF%d  GA%d\n",
			s.TeamName, s.Points, s.GoalDiff, s.GoalsFor, s.GoalsAgainst))
	}
	b.WriteString("\nTitle odds from 10,000 simulations:\n")
	for _, p := range input.Predictions {
		b.WriteString(fmt.Sprintf("  %s: %.1f%%\n", p.TeamName, p.WinProbability))
	}
	b.WriteString(fmt.Sprintf("\nMatches remaining: %d\n", input.RemainingMatches))
	return b.String()
}

// nextWeekPrompt builds the fixture preview prompt.
// Uses a different framing index from leaguePrompt so the two calls
// don't accidentally land on the same register.
func (g *GeminiGenerator) nextWeekPrompt(input Input) string {
	framings := []string{
		"You are previewing the upcoming fixtures for a betting preview column.",
		"You are a manager in a press conference asked about next week's games.",
		"You are a TV pundit asked which game you're most looking forward to and why.",
		"You are writing a pre-match briefing for a football analytics site.",
	}
	framing := framings[(time.Now().UnixNano()/1000)%int64(len(framings))]

	var b strings.Builder
	b.WriteString(framing + "\n")
	b.WriteString("Preview next week in 2–3 sentences. Name the games. Say which one is the key match and make one concrete score prediction.\n\n")

	b.WriteString("Fixtures:\n")
	for _, m := range input.NextWeekFixtures {
		b.WriteString(fmt.Sprintf("  %s (strength %.1f) vs %s (strength %.1f) — home side is %s\n",
			m.HomeTeam.Name, m.HomeTeam.SBase,
			m.AwayTeam.Name, m.AwayTeam.SBase,
			m.HomeTeam.Name,
		))
	}
	return b.String()
}

// ─── Gemini API types ─────────────────────────────────────────────────────────

type generationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type geminiRequest struct {
	Contents []struct {
		Parts []struct {
			Text string `json:"text"`
		} `json:"parts"`
	} `json:"contents"`
	GenerationConfig generationConfig `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (g *GeminiGenerator) generateText(ctx context.Context, prompt string) (string, error) {
	reqBody := geminiRequest{
		Contents: []struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		}{
			{Parts: []struct {
				Text string `json:"text"`
			}{{Text: prompt}}},
		},
		// temperature 0.85: high enough for varied phrasing,
		// low enough to stay factually grounded in the numbers.
		// maxOutputTokens keeps the output tight — without a cap
		// the model pads to meet a perceived length expectation.
		GenerationConfig: generationConfig{
			Temperature:     0.85,
			MaxOutputTokens: 160,
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(g.endpoint, g.model, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("gemini status: %s", resp.Status)
	}

	var parsed geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty gemini response")
	}
	return parsed.Candidates[0].Content.Parts[0].Text, nil
}
