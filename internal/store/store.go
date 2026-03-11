package store

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"time"

	"filecairn/internal/config"
)

type FileRecord struct {
	ID          string   `json:"id"`
	Path        string   `json:"path"`
	Name        string   `json:"name"`
	Ext         string   `json:"ext"`
	Size        int64    `json:"size"`
	ModUnix     int64    `json:"mod_unix"`
	Snippet     string   `json:"snippet"`
	Tokens      []string `json:"tokens"`
	Labels      []string `json:"labels"`
	ShareMode   string   `json:"share_mode"`
	AgentSource string   `json:"agent_source,omitempty"`
	IndexedAt   int64    `json:"indexed_at"`
}

type InvertedIndex map[string][]string

func dataPaths() (string, string, string, error) {
	h, err := config.HomeDir()
	if err != nil {
		return "", "", "", err
	}
	if err := os.MkdirAll(h, 0o755); err != nil {
		return "", "", "", err
	}
	return filepath.Join(h, "files.json"), filepath.Join(h, "inverted.json"), filepath.Join(h, "agent_created.jsonl"), nil
}

func LoadRecords() ([]FileRecord, error) {
	filePath, _, _, err := dataPaths()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return []FileRecord{}, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return []FileRecord{}, nil
	}
	var records []FileRecord
	if err := json.Unmarshal(b, &records); err != nil {
		return nil, err
	}
	return records, nil
}

func SaveRecords(records []FileRecord) error {
	filePath, _, _, err := dataPaths()
	if err != nil {
		return err
	}
	sort.Slice(records, func(i, j int) bool { return records[i].Path < records[j].Path })
	b, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, b, 0o644)
}

func LoadInverted() (InvertedIndex, error) {
	_, invPath, _, err := dataPaths()
	if err != nil {
		return nil, err
	}
	b, err := os.ReadFile(invPath)
	if err != nil {
		if os.IsNotExist(err) {
			return InvertedIndex{}, nil
		}
		return nil, err
	}
	if len(b) == 0 {
		return InvertedIndex{}, nil
	}
	idx := InvertedIndex{}
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	return idx, nil
}

func SaveInverted(idx InvertedIndex) error {
	_, invPath, _, err := dataPaths()
	if err != nil {
		return err
	}
	b, err := json.Marshal(idx)
	if err != nil {
		return err
	}
	return os.WriteFile(invPath, b, 0o644)
}

func BuildInverted(records []FileRecord) InvertedIndex {
	idx := InvertedIndex{}
	seen := map[string]map[string]bool{}
	for _, r := range records {
		for _, t := range r.Tokens {
			if _, ok := seen[t]; !ok {
				seen[t] = map[string]bool{}
			}
			if seen[t][r.ID] {
				continue
			}
			idx[t] = append(idx[t], r.ID)
			seen[t][r.ID] = true
		}
	}
	return idx
}

func RecordAgentCreated(path, agent, share string) error {
	if path == "" || agent == "" {
		return errors.New("path and agent are required")
	}
	_, _, acPath, err := dataPaths()
	if err != nil {
		return err
	}
	f, err := os.OpenFile(acPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	event := map[string]any{
		"timestamp":  time.Now().UTC().Format(time.RFC3339),
		"path":       path,
		"agent":      agent,
		"share_mode": share,
	}
	b, _ := json.Marshal(event)
	_, err = f.Write(append(b, '\n'))
	return err
}
