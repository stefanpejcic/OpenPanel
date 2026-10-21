package aiassistant

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"gist.github.com/stefanpejcic/openpanel/internal/core/tenantfs"
)

// only the end-user docs, admin/ and changelog/ are for server admins
var docPrefixes = []string{"panel/", "articles/"}

const (
	docsMaxAge     = 24 * time.Hour
	docsRetryAfter = 10 * time.Minute
	chunkMaxChars  = 3500
	maxChunks      = 4
	// the page the user is on only breaks ties, its words used to outrank the question ("add domain" asked on the dashboard got the dashboard docs)
	hintWeight = 0.25
	// chunks scoring below this share of the best match are noise the model mixes into its answer
	minRelScore = 0.5
)

var docsCachePath = filepath.Join(os.TempDir(), "openpanel-ai-docs.txt")

type chunk struct {
	Path  string
	Title string
	Text  string
	terms map[string]int
	size  int
}

type docIndex struct {
	chunks []chunk
	df     map[string]int
	avgLen float64
}

type docStore struct {
	mu        sync.Mutex
	index     *docIndex
	loadedAt  time.Time
	lastTry   time.Time
	url       string
	httpGet   func(ctx context.Context, url string) (string, error)
	cachePath string
}

var docs = &docStore{httpGet: fetchURL, cachePath: docsCachePath}

// get returns the index, refreshing it from the url once a day and falling back to the on-disk copy when the fetch fails
func (d *docStore) get(ctx context.Context, url string) *docIndex {
	d.mu.Lock()
	defer d.mu.Unlock()

	fresh := d.index != nil && d.url == url && time.Since(d.loadedAt) < docsMaxAge
	if fresh || (d.index != nil && time.Since(d.lastTry) < docsRetryAfter) {
		return d.index
	}

	if d.index == nil {
		if info, err := tenantfs.Stat(d.cachePath); err == nil && time.Since(info.ModTime()) < docsMaxAge {
			if raw, err := tenantfs.ReadFile(d.cachePath); err == nil {
				d.index, d.loadedAt, d.url = buildIndex(string(raw)), info.ModTime(), url
				return d.index
			}
		}
	}

	d.lastTry = time.Now()
	fetchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	raw, err := d.httpGet(fetchCtx, url)
	if err != nil {
		log.Printf("AI - fetching docs from %s failed: %v", url, err)
		if d.index == nil {
			if cached, err := tenantfs.ReadFile(d.cachePath); err == nil {
				d.index = buildIndex(string(cached))
			}
		}
		return d.index
	}

	filtered := filterDocs(raw)
	_ = tenantfs.WriteFile(d.cachePath, []byte(filtered), 0o600)
	d.index, d.loadedAt, d.url = buildIndex(filtered), time.Now(), url
	return d.index
}

func fetchURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status %d", resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 32<<20))
	return string(b), err
}

// filterDocs keeps only the "// File: panel/..." and "// File: articles/..." sections of llms-full.txt
func filterDocs(raw string) string {
	var out strings.Builder
	keep := false
	sc := bufio.NewScanner(strings.NewReader(raw))
	sc.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if path, ok := strings.CutPrefix(line, "// File: "); ok {
			keep = hasDocPrefix(strings.TrimSpace(path))
		}
		if keep {
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func hasDocPrefix(path string) bool {
	for _, p := range docPrefixes {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	return false
}

// buildIndex splits every doc into chunks at its ## headings and indexes them for bm25
func buildIndex(filtered string) *docIndex {
	idx := &docIndex{df: map[string]int{}}
	var path, title, heading string
	var body strings.Builder

	flush := func() {
		text := strings.TrimSpace(body.String())
		body.Reset()
		if path == "" || text == "" {
			return
		}
		for _, part := range splitLong(text, chunkMaxChars) {
			idx.add(path, title, heading, part)
		}
	}

	for _, line := range strings.Split(filtered, "\n") {
		if p, ok := strings.CutPrefix(line, "// File: "); ok {
			flush()
			path, title, heading = strings.TrimSpace(p), "", ""
			continue
		}
		if t, ok := strings.CutPrefix(line, "# "); ok && title == "" {
			title = strings.TrimSpace(t)
		} else if h, ok := strings.CutPrefix(line, "## "); ok {
			flush()
			heading = strings.TrimSpace(h)
		}
		body.WriteString(line)
		body.WriteByte('\n')
	}
	flush()

	total := 0
	for _, c := range idx.chunks {
		total += c.size
	}
	if len(idx.chunks) > 0 {
		idx.avgLen = float64(total) / float64(len(idx.chunks))
	}
	return idx
}

func (idx *docIndex) add(path, title, heading, text string) {
	terms := map[string]int{}
	size := 0
	for _, t := range tokenize(text) {
		terms[t]++
		size++
	}
	// path, title and heading say what the chunk is about, count them extra
	for _, t := range tokenize(strings.ReplaceAll(path, "/", " ") + " " + title + " " + heading) {
		terms[t] += 3
		size += 3
	}
	for t := range terms {
		idx.df[t]++
	}
	idx.chunks = append(idx.chunks, chunk{Path: path, Title: title, Text: text, terms: terms, size: size})
}

func splitLong(text string, max int) []string {
	if len(text) <= max {
		return []string{text}
	}
	var parts []string
	var cur strings.Builder
	for _, para := range strings.Split(text, "\n\n") {
		if cur.Len() > 0 && cur.Len()+len(para) > max {
			parts = append(parts, strings.TrimSpace(cur.String()))
			cur.Reset()
		}
		for len(para) > max {
			parts = append(parts, para[:max])
			para = para[max:]
		}
		cur.WriteString(para)
		cur.WriteString("\n\n")
	}
	if strings.TrimSpace(cur.String()) != "" {
		parts = append(parts, strings.TrimSpace(cur.String()))
	}
	return parts
}

var hostnameRE = regexp.MustCompile(`(?i)\b[a-z0-9-]+(\.[a-z0-9-]+)*\.[a-z]{2,}\b`)

var stopwords = map[string]bool{
	"the": true, "and": true, "for": true, "how": true, "can": true, "you": true, "your": true, "are": true,
	"what": true, "with": true, "this": true, "that": true, "from": true, "have": true, "not": true, "but": true,
	"does": true, "why": true, "who": true, "was": true, "will": true, "into": true, "there": true, "should": true,
	"would": true, "could": true, "about": true, "then": true, "than": true, "them": true, "they": true, "its": true,
	"use": true, "using": true, "get": true, "want": true, "need": true, "please": true, "doc": true, "docs": true,
	"do": true, "is": true, "it": true, "to": true, "in": true, "on": true, "of": true, "my": true, "me": true,
	"an": true, "or": true, "be": true, "if": true, "at": true, "by": true, "as": true, "so": true, "we": true,
}

func tokenize(s string) []string {
	words := strings.FieldsFunc(strings.ToLower(s), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	out := words[:0]
	for _, w := range words {
		if len(w) < 2 || stopwords[w] {
			continue
		}
		out = append(out, stem(w))
	}
	return out
}

// stem is a very rough plural/verb trim so "domains" matches "domain" and "installing" matches "install"
func stem(w string) string {
	for _, suf := range []string{"ing", "es", "s"} {
		if len(w) > len(suf)+3 && strings.HasSuffix(w, suf) {
			return w[:len(w)-len(suf)]
		}
	}
	return w
}

// search ranks chunks by bm25 and returns the best ones that fit in budget characters, the same query always gives the same chunks so the provider's prompt cache keeps working
func (idx *docIndex) search(query, hint string, budget int) []chunk {
	if idx == nil || len(idx.chunks) == 0 {
		return nil
	}
	qterms := map[string]float64{}
	for _, t := range tokenize(hint) {
		qterms[t] = hintWeight
	}
	// the user's own domain names say nothing about the topic, and their tld ("rs") is a rare word that pulls in random articles
	for _, t := range tokenize(hostnameRE.ReplaceAllString(query, " ")) {
		qterms[t] = 1
	}
	if len(qterms) == 0 {
		return nil
	}

	const k1, b = 1.2, 0.75
	n := float64(len(idx.chunks))
	type scored struct {
		i     int
		score float64
	}
	var results []scored
	for i, c := range idx.chunks {
		score := 0.0
		for t, w := range qterms {
			tf := float64(c.terms[t])
			if tf == 0 {
				continue
			}
			df := float64(idx.df[t])
			idf := math.Log(1 + (n-df+0.5)/(df+0.5))
			score += w * idf * tf * (k1 + 1) / (tf + k1*(1-b+b*float64(c.size)/idx.avgLen))
		}
		// panel/ pages describe the ui the user is looking at, articles are more general
		if strings.HasPrefix(c.Path, "panel/") {
			score *= 1.5
		}
		if score > 0 {
			results = append(results, scored{i, score})
		}
	}
	sort.SliceStable(results, func(a, b int) bool { return results[a].score > results[b].score })

	var picked []chunk
	used := 0
	for _, r := range results {
		if r.score < results[0].score*minRelScore {
			break
		}
		c := idx.chunks[r.i]
		if used+len(c.Text) > budget {
			continue
		}
		picked = append(picked, c)
		used += len(c.Text)
		if len(picked) >= maxChunks {
			break
		}
	}
	return picked
}
