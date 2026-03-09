package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"fileatlas/internal/config"
	"fileatlas/internal/indexer"
	"fileatlas/internal/store"
	"fileatlas/internal/util"
)

func RunAndPersistScan(cfg config.Config, roots []string) (indexer.Stats, int, error) {
	existing, err := store.LoadRecords()
	if err != nil {
		return indexer.Stats{}, 0, err
	}
	records, stats := indexer.RunScan(cfg, roots, existing)
	if err := store.SaveRecords(records); err != nil {
		return stats, 0, err
	}
	inv := store.BuildInverted(records)
	if err := store.SaveInverted(inv); err != nil {
		return stats, 0, err
	}
	return stats, len(records), nil
}

func RegisterCreatedFile(cfg config.Config, path, agent, share string) (store.FileRecord, error) {
	if path == "" {
		return store.FileRecord{}, errors.New("path is required")
	}
	if agent == "" {
		agent = "unknown"
	}
	if share == "" {
		share = "full"
	}
	path = util.NormalizePath(path)
	fi, err := os.Stat(path)
	if err != nil {
		return store.FileRecord{}, err
	}
	if !fi.Mode().IsRegular() {
		return store.FileRecord{}, fmt.Errorf("path is not a regular file: %s", path)
	}

	records, err := store.LoadRecords()
	if err != nil {
		return store.FileRecord{}, err
	}

	mod := fi.ModTime().Unix()
	sz := fi.Size()
	rec := store.FileRecord{
		ID:          util.HashID(path, mod, sz),
		Path:        path,
		Name:        filepath.Base(path),
		Ext:         strings.ToLower(filepath.Ext(path)),
		Size:        sz,
		ModUnix:     mod,
		ShareMode:   share,
		AgentSource: agent,
		IndexedAt:   time.Now().Unix(),
	}

	textForTokens := rec.Name + " " + rec.Path
	allowContentRead := cfg.ContentReadEnabled || rec.ShareMode == "full"
	if allowContentRead && util.IsLikelyTextFile(path) {
		rec.Snippet = util.ReadSnippet(path, cfg.MaxFileBytes)
		textForTokens += " " + rec.Snippet
	}
	rec.Tokens = util.Tokenize(textForTokens)
	rec.Labels = util.LabelsFor(path, rec.Tokens)

	updated := false
	for i := range records {
		if records[i].Path == path {
			records[i] = rec
			updated = true
			break
		}
	}
	if !updated {
		records = append(records, rec)
	}

	if err := store.SaveRecords(records); err != nil {
		return store.FileRecord{}, err
	}
	if err := store.SaveInverted(store.BuildInverted(records)); err != nil {
		return store.FileRecord{}, err
	}
	if err := store.RecordAgentCreated(path, agent, share); err != nil {
		return store.FileRecord{}, err
	}
	return rec, nil
}
