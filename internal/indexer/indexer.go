package indexer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fileatlas/internal/config"
	"fileatlas/internal/store"
	"fileatlas/internal/util"
)

type Stats struct {
	ScannedFiles int64
	IndexedFiles int64
	SkippedFiles int64
	Errors       int64
	Duration     time.Duration
}

func RunScan(cfg config.Config, roots []string, existing []store.FileRecord) ([]store.FileRecord, Stats) {
	start := time.Now()
	if len(roots) == 0 {
		roots = cfg.ScanRoots
	}

	existingMap := map[string]store.FileRecord{}
	for _, r := range existing {
		existingMap[r.Path] = r
	}

	fileCh := make(chan string, 2048)
	recCh := make(chan store.FileRecord, 2048)
	errCh := make(chan error, 256)

	var scanned, indexed, skipped, errorsCount int64
	var wg sync.WaitGroup
	workerCount := runtime.NumCPU()
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range fileCh {
				atomic.AddInt64(&scanned, 1)
				fi, err := os.Stat(p)
				if err != nil {
					atomic.AddInt64(&errorsCount, 1)
					errCh <- err
					continue
				}
				mod := fi.ModTime().Unix()
				sz := fi.Size()

				if ex, ok := existingMap[p]; ok && ex.ModUnix == mod && ex.Size == sz {
					recCh <- ex
					atomic.AddInt64(&skipped, 1)
					continue
				}

				rec := store.FileRecord{
					ID:        util.HashID(p, mod, sz),
					Path:      p,
					Name:      filepath.Base(p),
					Ext:       strings.ToLower(filepath.Ext(p)),
					Size:      sz,
					ModUnix:   mod,
					ShareMode: "private",
					IndexedAt: time.Now().Unix(),
				}

				if old, ok := existingMap[p]; ok {
					rec.ShareMode = old.ShareMode
					rec.AgentSource = old.AgentSource
				}

				textForTokens := rec.Name + " " + rec.Path
				allowContentRead := cfg.ContentReadEnabled || rec.ShareMode == "full"
				if allowContentRead && util.IsLikelyTextFile(p) {
					rec.Snippet = util.ReadSnippet(p, cfg.MaxFileBytes)
					textForTokens += " " + rec.Snippet
				}
				rec.Tokens = util.Tokenize(textForTokens)
				rec.Labels = util.LabelsFor(p, rec.Tokens)
				recCh <- rec
				atomic.AddInt64(&indexed, 1)
			}
		}()
	}

	go func() {
		for _, root := range roots {
			root = util.NormalizePath(root)
			filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					atomic.AddInt64(&errorsCount, 1)
					errCh <- err
					return nil
				}
				if d.IsDir() {
					if shouldSkipDir(path, cfg.IgnoreDirs) {
						return filepath.SkipDir
					}
					return nil
				}
				if !d.Type().IsRegular() {
					atomic.AddInt64(&skipped, 1)
					return nil
				}
				fileCh <- util.NormalizePath(path)
				return nil
			})
		}
		close(fileCh)
		wg.Wait()
		close(recCh)
		close(errCh)
	}()

	out := []store.FileRecord{}
	seen := map[string]bool{}
	for rec := range recCh {
		if seen[rec.Path] {
			continue
		}
		seen[rec.Path] = true
		out = append(out, rec)
	}

	// Drain errors channel so goroutine can finish cleanly.
	for range errCh {
	}

	stats := Stats{
		ScannedFiles: atomic.LoadInt64(&scanned),
		IndexedFiles: atomic.LoadInt64(&indexed),
		SkippedFiles: atomic.LoadInt64(&skipped),
		Errors:       atomic.LoadInt64(&errorsCount),
		Duration:     time.Since(start),
	}

	fmt.Printf("Scan complete: scanned=%d indexed=%d skipped=%d errors=%d duration=%s\n",
		stats.ScannedFiles, stats.IndexedFiles, stats.SkippedFiles, stats.Errors, stats.Duration.Round(time.Millisecond))

	return out, stats
}

func shouldSkipDir(path string, ignore []string) bool {
	p := strings.ToLower(path)
	for _, ig := range ignore {
		ig = strings.ToLower(strings.TrimSpace(ig))
		if ig == "" {
			continue
		}
		if strings.Contains(p, strings.ToLower(string(filepath.Separator)+ig+string(filepath.Separator))) {
			return true
		}
		if strings.HasSuffix(p, string(filepath.Separator)+ig) {
			return true
		}
	}
	return false
}
