package util

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var tokenRe = regexp.MustCompile(`[A-Za-z0-9_]{2,}`)

var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "that": true,
	"from": true, "this": true, "file": true, "you": true, "your": true,
	"into": true, "have": true, "are": true, "was": true, "not": true,
	"but": true, "all": true, "can": true, "use": true, "using": true,
	"http": true, "https": true, "com": true,
}

func NormalizePath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(abs)
}

func HashID(path string, modUnix int64, size int64) string {
	h := sha1.New()
	io.WriteString(h, fmt.Sprintf("%s|%d|%d", path, modUnix, size))
	return hex.EncodeToString(h.Sum(nil))
}

func Tokenize(text string) []string {
	matches := tokenRe.FindAllString(strings.ToLower(text), -1)
	if len(matches) == 0 {
		return []string{}
	}
	freq := map[string]int{}
	for _, m := range matches {
		if stopWords[m] {
			continue
		}
		freq[m]++
	}
	type kv struct {
		K string
		V int
	}
	arr := make([]kv, 0, len(freq))
	for k, v := range freq {
		arr = append(arr, kv{K: k, V: v})
	}
	sort.Slice(arr, func(i, j int) bool {
		if arr[i].V == arr[j].V {
			return arr[i].K < arr[j].K
		}
		return arr[i].V > arr[j].V
	})
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		out = append(out, e.K)
		if len(out) >= 128 {
			break
		}
	}
	return out
}

func IsLikelyTextFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".txt", ".md", ".markdown", ".go", ".py", ".js", ".ts", ".tsx", ".jsx", ".json", ".yaml", ".yml", ".toml", ".ini", ".cfg", ".conf", ".java", ".c", ".cpp", ".h", ".hpp", ".rs", ".swift", ".kt", ".sh", ".zsh", ".bash", ".csv", ".sql", ".log", ".xml", ".html", ".css":
		return true
	default:
		return false
	}
}

func ReadSnippet(path string, maxBytes int64) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	r := bufio.NewReader(f)
	buf := make([]byte, maxBytes)
	n, _ := r.Read(buf)
	if n <= 0 {
		return ""
	}
	s := string(buf[:n])
	s = strings.ReplaceAll(s, "\x00", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.Join(strings.Fields(s), " ")
	if len(s) > 500 {
		s = s[:500]
	}
	return s
}

func LabelsFor(path string, tokens []string) []string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(path)), ".")
	labels := []string{}
	if ext != "" {
		labels = append(labels, "ext:"+ext)
	}
	name := strings.ToLower(filepath.Base(path))
	if strings.Contains(path, "/desktop/") {
		labels = append(labels, "folder:desktop")
	}
	if strings.Contains(path, "/documents/") {
		labels = append(labels, "folder:documents")
	}
	if strings.Contains(path, "/downloads/") {
		labels = append(labels, "folder:downloads")
	}
	if strings.Contains(name, "readme") {
		labels = append(labels, "doc:readme")
	}
	for i, t := range tokens {
		labels = append(labels, "topic:"+t)
		if i >= 4 {
			break
		}
	}
	return labels
}
