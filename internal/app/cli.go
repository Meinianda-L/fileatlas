package app

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"fileatlas/internal/api"
	"fileatlas/internal/config"
	"fileatlas/internal/core"
	"fileatlas/internal/search"
	"fileatlas/internal/store"
	"fileatlas/internal/util"
)

func Run(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}
	cmd := args[0]
	switch cmd {
	case "init":
		return runInit()
	case "scan":
		return runScan(args[1:])
	case "find":
		return runFind(args[1:])
	case "register-created":
		return runRegisterCreated(args[1:])
	case "serve":
		return runServe(args[1:])
	case "status":
		return runStatus()
	case "content":
		return runContent(args[1:])
	case "help", "-h", "--help":
		printHelp()
		return nil
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}

func printHelp() {
	fmt.Println("FileAtlas")
	fmt.Println("Fast local index and search for your files")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  fileatlas <command> [flags]")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  init             Interactive setup (providers, privacy, roots)")
	fmt.Println("  scan             Build or refresh the index")
	fmt.Println("  find             Search indexed files by keywords")
	fmt.Println("  register-created Register a file written by an agent/tool")
	fmt.Println("  serve            Start the local HTTP API")
	fmt.Println("  status           Show config and index summary")
	fmt.Println("  content          Toggle content indexing (on|off)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  fileatlas init")
	fmt.Println("  fileatlas scan --roots ~/Documents,~/Desktop")
	fmt.Println("  fileatlas find \"quarterly budget\"")
	fmt.Println("  fileatlas register-created --path /tmp/notes.md --agent openclaw --share full")
}

func runInit() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	cfg := config.Default(home)
	r := bufio.NewReader(os.Stdin)

	fmt.Println("=== FileAtlas Setup ===")
	fmt.Println("Provider configuration")
	fmt.Println("Add one or more model providers (endpoint + model + api key env var).")

	for {
		addProvider, err := askYesNo(r, "Add a model provider now?", false)
		if err != nil {
			return err
		}
		if !addProvider {
			break
		}
		pname, err := askText(r, "Provider name (example: openai, ollama, local-http)", "")
		if err != nil {
			return err
		}
		endpoint, err := askText(r, "Endpoint URL (example: https://api.openai.com/v1)", "")
		if err != nil {
			return err
		}
		model, err := askText(r, "Default model name", "")
		if err != nil {
			return err
		}
		apiKeyEnv, err := askText(r, "API key env var name (example: PROVIDER_API_KEY)", "")
		if err != nil {
			return err
		}
		cfg.Providers = append(cfg.Providers, config.Provider{
			Name:      strings.TrimSpace(pname),
			Endpoint:  strings.TrimSpace(endpoint),
			Model:     strings.TrimSpace(model),
			APIKeyEnv: strings.TrimSpace(apiKeyEnv),
		})
	}
	if len(cfg.Providers) > 0 {
		defaultActive := cfg.Providers[0].Name
		active, err := askText(r, "Active provider name", defaultActive)
		if err != nil {
			return err
		}
		cfg.ActiveProvider = active
	}

	allowContent, err := askYesNo(r, "Allow content indexing now? (default: no)", false)
	if err != nil {
		return err
	}
	cfg.ContentReadEnabled = allowContent

	fmt.Println("")
	allowAll, err := askYesNo(r, "Scan your full home directory now?", false)
	if err != nil {
		return err
	}
	if allowAll {
		cfg.ScanRoots = []string{home}
	} else {
		rootsRaw, err := askText(r, "Enter scan roots (comma-separated)", strings.Join(cfg.ScanRoots, ","))
		if err != nil {
			return err
		}
		cfg.ScanRoots = parseRoots(rootsRaw)
	}

	if err := config.Save(cfg); err != nil {
		return err
	}
	cfgPath, _ := config.ConfigPath()
	fmt.Printf("Saved config: %s\n", cfgPath)

	doScan, err := askYesNo(r, "Build index now?", true)
	if err != nil {
		return err
	}
	if doScan {
		stats, total, err := core.RunAndPersistScan(cfg, cfg.ScanRoots)
		if err != nil {
			return err
		}
		fmt.Printf("Index ready. files=%d scanned=%d indexed=%d skipped=%d errors=%d\n",
			total, stats.ScannedFiles, stats.IndexedFiles, stats.SkippedFiles, stats.Errors)
	}
	return nil
}

func runScan(args []string) error {
	cfg, err := config.Require()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	all := fs.Bool("all", false, "scan full home directory")
	rootsCSV := fs.String("roots", "", "comma-separated roots")
	if err := fs.Parse(args); err != nil {
		return err
	}

	roots := cfg.ScanRoots
	if *all {
		ok, err := confirmOnce("You requested a full-home scan. Continue?")
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Scan cancelled.")
			return nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		roots = []string{home}
	}
	if strings.TrimSpace(*rootsCSV) != "" {
		roots = parseRoots(*rootsCSV)
	}
	if len(roots) == 0 {
		return errors.New("no scan roots configured; run `fileatlas init` or pass --roots")
	}

	stats, total, err := core.RunAndPersistScan(cfg, roots)
	if err != nil {
		return err
	}
	fmt.Printf("Scan persisted. files=%d scanned=%d indexed=%d skipped=%d errors=%d\n",
		total, stats.ScannedFiles, stats.IndexedFiles, stats.SkippedFiles, stats.Errors)
	return nil
}

func confirmOnce(prompt string) (bool, error) {
	r := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("%s [y/N]: ", prompt)
		line, err := r.ReadString('\n')
		if err != nil {
			return false, err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" || line == "n" || line == "no" {
			return false, nil
		}
		if line == "y" || line == "yes" {
			return true, nil
		}
	}
}

func runFind(args []string) error {
	fs := flag.NewFlagSet("find", flag.ContinueOnError)
	limit := fs.Int("limit", 20, "max results")
	if err := fs.Parse(args); err != nil {
		return err
	}
	query := strings.TrimSpace(strings.Join(fs.Args(), " "))
	if query == "" {
		return errors.New("query is required; usage: fileatlas find [--limit 20] <query>")
	}
	records, err := store.LoadRecords()
	if err != nil {
		return err
	}
	if len(records) == 0 {
		fmt.Println("Index is empty. Run `fileatlas scan` first.")
		return nil
	}
	idx, err := store.LoadInverted()
	if err != nil {
		return err
	}
	results := search.Find(records, idx, query, *limit)
	if len(results) == 0 {
		fmt.Println("No results")
		return nil
	}
	for i, res := range results {
		fmt.Printf("%d. %s\n", i+1, res.Record.Path)
		fmt.Printf("   score=%.3f share=%s labels=%s\n", res.Score, res.Record.ShareMode, strings.Join(res.Record.Labels, ","))
		if res.Record.Snippet != "" {
			fmt.Printf("   snippet=%q\n", trimDisplay(res.Record.Snippet, 120))
		}
		fmt.Printf("   why=%s\n", strings.Join(res.Why, " | "))
	}
	return nil
}

func runRegisterCreated(args []string) error {
	cfg, err := config.Require()
	if err != nil {
		return err
	}
	fs := flag.NewFlagSet("register-created", flag.ContinueOnError)
	path := fs.String("path", "", "file path")
	agent := fs.String("agent", "", "agent name")
	share := fs.String("share", "full", "share mode: private|summary|full")
	if err := fs.Parse(args); err != nil {
		return err
	}
	rec, err := core.RegisterCreatedFile(cfg, *path, *agent, *share)
	if err != nil {
		return err
	}
	fmt.Printf("Registered: %s\n", rec.Path)
	fmt.Printf("share=%s agent=%s labels=%s\n", rec.ShareMode, rec.AgentSource, strings.Join(rec.Labels, ","))
	return nil
}

func runServe(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	addr := fs.String("addr", "127.0.0.1:4819", "listen address")
	if err := fs.Parse(args); err != nil {
		return err
	}
	return api.Start(*addr)
}

func runStatus() error {
	cfg, err := config.Require()
	if err != nil {
		return err
	}
	records, err := store.LoadRecords()
	if err != nil {
		return err
	}
	home, _ := config.HomeDir()
	cfgPath, _ := config.ConfigPath()
	fmt.Println("FileAtlas status")
	fmt.Printf("  config: %s\n", cfgPath)
	fmt.Printf("  data dir: %s\n", home)
	fmt.Printf("  roots: %s\n", strings.Join(cfg.ScanRoots, ", "))
	fmt.Printf("  content_read_enabled: %v\n", cfg.ContentReadEnabled)
	fmt.Printf("  providers: %d active=%s\n", len(cfg.Providers), cfg.ActiveProvider)
	fmt.Printf("  indexed files: %d\n", len(records))
	return nil
}

func runContent(args []string) error {
	if len(args) != 1 {
		return errors.New("usage: fileatlas content on|off")
	}
	cfg, err := config.Require()
	if err != nil {
		return err
	}
	switch strings.ToLower(args[0]) {
	case "on":
		cfg.ContentReadEnabled = true
	case "off":
		cfg.ContentReadEnabled = false
	default:
		return errors.New("usage: fileatlas content on|off")
	}
	if err := config.Save(cfg); err != nil {
		return err
	}
	fmt.Printf("Content indexing set to %v\n", cfg.ContentReadEnabled)
	return nil
}

func askYesNo(r *bufio.Reader, prompt string, def bool) (bool, error) {
	defStr := "y/N"
	if def {
		defStr = "Y/n"
	}
	for {
		fmt.Printf("%s [%s]: ", prompt, defStr)
		line, err := r.ReadString('\n')
		if err != nil {
			return false, err
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "" {
			return def, nil
		}
		switch line {
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		}
	}
}

func askText(r *bufio.Reader, prompt, def string) (string, error) {
	if def != "" {
		fmt.Printf("%s [%s]: ", prompt, def)
	} else {
		fmt.Printf("%s: ", prompt)
	}
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return def, nil
	}
	return line, nil
}

func parseRoots(csv string) []string {
	parts := strings.Split(csv, ",")
	roots := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if strings.HasPrefix(p, "~") {
			h, _ := os.UserHomeDir()
			p = filepath.Join(h, strings.TrimPrefix(p, "~/"))
		}
		roots = append(roots, util.NormalizePath(p))
	}
	return roots
}

func trimDisplay(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
