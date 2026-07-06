// Package index provides a small in-memory inverted index for full-text search
// over documents identified by an opaque string ID.
//
// It tokenizes text into lower-cased alphanumeric terms, records term
// frequencies per document, and ranks results with a tf-idf score. It is
// concurrency-safe and holds everything in memory; it is meant for modest
// corpora (thousands of documents), not as a replacement for a real search
// engine.
package index

import (
	"math"
	"sort"
	"strings"
	"sync"
	"unicode"
)

// Index is a concurrency-safe inverted index. The zero value is not usable;
// call New.
type Index struct {
	mu       sync.RWMutex
	postings map[string]map[string]int // term -> docID -> term frequency
	lengths  map[string]int            // docID -> total term count
}

// New returns an empty index.
func New() *Index {
	return &Index{
		postings: make(map[string]map[string]int),
		lengths:  make(map[string]int),
	}
}

// Add indexes text under id, replacing any previous content for that id.
func (ix *Index) Add(id, text string) {
	terms := Tokenize(text)
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.removeLocked(id)
	if len(terms) == 0 {
		return
	}
	ix.lengths[id] = len(terms)
	for _, term := range terms {
		p := ix.postings[term]
		if p == nil {
			p = make(map[string]int)
			ix.postings[term] = p
		}
		p[id]++
	}
}

// Delete removes the document with the given id from the index.
func (ix *Index) Delete(id string) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.removeLocked(id)
}

func (ix *Index) removeLocked(id string) {
	if _, ok := ix.lengths[id]; !ok {
		return
	}
	delete(ix.lengths, id)
	for term, p := range ix.postings {
		if _, ok := p[id]; ok {
			delete(p, id)
			if len(p) == 0 {
				delete(ix.postings, term)
			}
		}
	}
}

// Len returns the number of indexed documents.
func (ix *Index) Len() int {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return len(ix.lengths)
}

// Result is a scored search hit.
type Result struct {
	ID    string
	Score float64
}

// Search tokenizes query and returns matching documents ranked by descending
// tf-idf score. Documents containing more of the query terms rank higher.
// Results are limited to limit hits; a limit <= 0 returns all matches.
func (ix *Index) Search(query string, limit int) []Result {
	terms := Tokenize(query)
	if len(terms) == 0 {
		return nil
	}

	ix.mu.RLock()
	defer ix.mu.RUnlock()

	n := float64(len(ix.lengths))
	if n == 0 {
		return nil
	}

	scores := make(map[string]float64)
	for _, term := range dedupe(terms) {
		p := ix.postings[term]
		if len(p) == 0 {
			continue
		}
		// Smoothed inverse document frequency.
		idf := math.Log(1 + n/float64(len(p)))
		for id, tf := range p {
			norm := float64(ix.lengths[id])
			if norm == 0 {
				norm = 1
			}
			scores[id] += (float64(tf) / norm) * idf
		}
	}

	results := make([]Result, 0, len(scores))
	for id, s := range scores {
		results = append(results, Result{ID: id, Score: s})
	}
	sort.Slice(results, func(i, j int) bool {
		if results[i].Score != results[j].Score {
			return results[i].Score > results[j].Score
		}
		return results[i].ID < results[j].ID // stable tie-break
	})
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results
}

// Tokenize splits text into lower-cased tokens of letters and digits. It is
// exported so callers can index and query consistently.
func Tokenize(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func dedupe(terms []string) []string {
	seen := make(map[string]struct{}, len(terms))
	out := terms[:0:0]
	for _, t := range terms {
		if _, ok := seen[t]; ok {
			continue
		}
		seen[t] = struct{}{}
		out = append(out, t)
	}
	return out
}
