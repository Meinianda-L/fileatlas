package search

import (
	"math"
	"sort"
	"strings"
	"time"

	"fileatlas/internal/store"
	"fileatlas/internal/util"
)

type Result struct {
	Record      store.FileRecord `json:"record"`
	Score       float64          `json:"score"`
	MatchTokens []string         `json:"match_tokens"`
	Why         []string         `json:"why"`
}

func Find(records []store.FileRecord, inverted store.InvertedIndex, query string, limit int) []Result {
	if limit <= 0 {
		limit = 20
	}
	queryTokens := util.Tokenize(query)
	if len(queryTokens) == 0 {
		return []Result{}
	}

	byID := map[string]store.FileRecord{}
	for _, r := range records {
		byID[r.ID] = r
	}

	candidates := map[string]*Result{}
	for _, qt := range queryTokens {
		ids := inverted[qt]
		for _, id := range ids {
			r, ok := byID[id]
			if !ok {
				continue
			}
			entry, ok := candidates[id]
			if !ok {
				entry = &Result{Record: r, Score: 0.0, MatchTokens: []string{}, Why: []string{}}
				candidates[id] = entry
			}
			tf := tokenFrequency(r.Tokens, qt)
			entry.Score += 0.45 * float64(tf)
			if !contains(entry.MatchTokens, qt) {
				entry.MatchTokens = append(entry.MatchTokens, qt)
			}
		}
	}

	results := make([]Result, 0, len(candidates))
	now := time.Now().Unix()
	for _, r := range candidates {
		ageDays := float64(now-r.Record.ModUnix) / 86400.0
		recency := math.Exp(-ageDays / 45.0)
		pathBoost := pathPriorityBoost(strings.ToLower(r.Record.Path))
		labelBoost := labelMatchBoost(r.Record.Labels, queryTokens)
		r.Score += 0.30*recency + 0.15*pathBoost + 0.10*labelBoost
		r.Why = append(r.Why,
			"matched tokens: "+strings.Join(r.MatchTokens, ", "),
			"labels: "+strings.Join(r.Record.Labels, ", "),
		)
		results = append(results, *r)
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].Score == results[j].Score {
			return results[i].Record.ModUnix > results[j].Record.ModUnix
		}
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results
}

func tokenFrequency(tokens []string, token string) int {
	count := 0
	for _, t := range tokens {
		if t == token {
			count++
		}
	}
	if count == 0 {
		return 1
	}
	return count
}

func contains(arr []string, item string) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}

func pathPriorityBoost(path string) float64 {
	score := 0.0
	if strings.Contains(path, "/desktop/") {
		score += 0.9
	}
	if strings.Contains(path, "/documents/") {
		score += 0.7
	}
	if strings.Contains(path, "/downloads/") {
		score += 0.5
	}
	return score
}

func labelMatchBoost(labels []string, queryTokens []string) float64 {
	if len(labels) == 0 {
		return 0
	}
	match := 0
	for _, l := range labels {
		for _, q := range queryTokens {
			if strings.Contains(l, q) {
				match++
				break
			}
		}
	}
	return float64(match) / float64(len(labels))
}
