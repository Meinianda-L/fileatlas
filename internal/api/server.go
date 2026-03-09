package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"fileatlas/internal/config"
	"fileatlas/internal/core"
	"fileatlas/internal/search"
	"fileatlas/internal/store"
)

type findRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

type scanRequest struct {
	All   bool     `json:"all"`
	Roots []string `json:"roots"`
}

type registerRequest struct {
	Path  string `json:"path"`
	Agent string `json:"agent"`
	Share string `json:"share"`
}

func Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"ok": true, "service": "fileatlas"}, http.StatusOK)
	})

	mux.HandleFunc("/v1/status", func(w http.ResponseWriter, r *http.Request) {
		cfg, err := config.Require()
		if err != nil {
			writeErr(w, err, http.StatusBadRequest)
			return
		}
		records, err := store.LoadRecords()
		if err != nil {
			writeErr(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{
			"roots":                cfg.ScanRoots,
			"content_read_enabled": cfg.ContentReadEnabled,
			"active_provider":      cfg.ActiveProvider,
			"providers":            cfg.Providers,
			"indexed_files":        len(records),
		}, http.StatusOK)
	})

	mux.HandleFunc("/v1/find", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, fmt.Errorf("method not allowed"), http.StatusMethodNotAllowed)
			return
		}
		var req findRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, err, http.StatusBadRequest)
			return
		}
		records, err := store.LoadRecords()
		if err != nil {
			writeErr(w, err, http.StatusInternalServerError)
			return
		}
		inv, err := store.LoadInverted()
		if err != nil {
			writeErr(w, err, http.StatusInternalServerError)
			return
		}
		results := search.Find(records, inv, req.Query, req.Limit)
		writeJSON(w, map[string]any{"results": results}, http.StatusOK)
	})

	mux.HandleFunc("/v1/scan", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, fmt.Errorf("method not allowed"), http.StatusMethodNotAllowed)
			return
		}
		cfg, err := config.Require()
		if err != nil {
			writeErr(w, err, http.StatusBadRequest)
			return
		}
		var req scanRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, err, http.StatusBadRequest)
			return
		}
		roots := req.Roots
		if req.All {
			home, err := os.UserHomeDir()
			if err != nil {
				writeErr(w, err, http.StatusInternalServerError)
				return
			}
			roots = []string{home}
		}
		stats, total, err := core.RunAndPersistScan(cfg, roots)
		if err != nil {
			writeErr(w, err, http.StatusInternalServerError)
			return
		}
		writeJSON(w, map[string]any{"stats": stats, "total_files": total}, http.StatusOK)
	})

	mux.HandleFunc("/v1/register-created", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeErr(w, fmt.Errorf("method not allowed"), http.StatusMethodNotAllowed)
			return
		}
		cfg, err := config.Require()
		if err != nil {
			writeErr(w, err, http.StatusBadRequest)
			return
		}
		var req registerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErr(w, err, http.StatusBadRequest)
			return
		}
		rec, err := core.RegisterCreatedFile(cfg, req.Path, req.Agent, req.Share)
		if err != nil {
			writeErr(w, err, http.StatusBadRequest)
			return
		}
		writeJSON(w, map[string]any{"record": rec}, http.StatusOK)
	})

	fmt.Printf("FileAtlas API listening on http://%s\n", addr)
	return http.ListenAndServe(addr, mux)
}

func writeJSON(w http.ResponseWriter, payload any, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeErr(w http.ResponseWriter, err error, code int) {
	writeJSON(w, map[string]any{"error": err.Error()}, code)
}
