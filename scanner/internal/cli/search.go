// Package cli implements the VibeShield v2 console: a searchable,
// keyboard-driven terminal UI plus the non-interactive `vibeshield search`
// command that shares its ranking. Everything here is static and offline —
// no network calls, no code leaves the machine (CLAUDE.md hard rule 2).
package cli

import (
	"sort"
	"strings"
	"unicode"
)

// Kind tells the console what to do with a selected entry.
type Kind string

const (
	// KindAction exits the console and re-runs the CLI with Args.
	KindAction Kind = "action"
	// KindInfo opens a read-only detail pane (rule docs, agent recipes).
	KindInfo Kind = "info"
)

// Item is one searchable entry in the console.
type Item struct {
	Kind     Kind
	Group    string
	Title    string
	Summary  string
	Keywords []string
	Args     []string
	Body     string
}

// field is one weighted haystack the query is scored against. Weights mirror
// how a human reads a row: the title dominates, keywords are strong signal,
// the one-line summary is a tiebreaker.
type field struct {
	text    string
	weight  int
	keyword bool // keywords reinforce a match instead of competing with it
}

func (it Item) fields() []field {
	fs := make([]field, 0, 3+len(it.Keywords))
	fs = append(fs,
		field{text: it.Title, weight: 12},
		field{text: it.Group, weight: 4},
		field{text: it.Summary, weight: 3},
	)
	for _, k := range it.Keywords {
		fs = append(fs, field{text: k, weight: 6, keyword: true})
	}
	return fs
}

// normalize lowercases s, folds punctuation to spaces and collapses runs of
// whitespace. Folding punctuation lets "claude-code" match "claude code" and
// "pre_commit" match "pre-commit" — the same spelling drift agents introduce.
func normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToLower(s) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '+' {
			b.WriteRune(r)
			continue
		}
		b.WriteRune(' ')
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// tokenize splits a query into normalized tokens.
func tokenize(q string) []string {
	return strings.Fields(normalize(q))
}

// scoreText rates how well a single token matches text. 0 means no match.
// Bigger is better.
func scoreText(q, text string) int {
	text = normalize(text)
	if text == "" || q == "" {
		return 0
	}
	switch {
	case text == q:
		return 100
	case strings.HasPrefix(text, q):
		return 80
	case strings.Contains(text, " "+q):
		return 70 // matches at a word boundary
	case strings.Contains(text, q):
		return 50
	}
	if bonus, ok := subsequence(q, text); ok {
		return bonus
	}
	return 0
}

// minFuzzyRunes is the shortest token allowed to fall back to subsequence
// matching. Below it the fuzzy path is pure noise rather than recall: "aws"
// subsequence-matches "CORS allows all origins" (a-w-s) and "sec" matches
// almost any sentence, so a three-letter search would return the whole catalog.
const minFuzzyRunes = 4

// maxFuzzyGaps bounds how scattered a subsequence match may be, as a multiple
// of the query length. "vs017" inside "vs sec 017" skips 5 runes, which is a
// real abbreviation; a four-letter query spread across a 60-rune sentence is
// not a match, it is coincidence.
const maxFuzzyGaps = 4

// subsequence reports whether every rune of q appears in text in order, and
// returns a compactness bonus — fewer skipped runes score higher. This is the
// fuzzy fallback, so "vssec017" still finds "VS-SEC-017".
func subsequence(q, text string) (int, bool) {
	qr, tr := []rune(q), []rune(text)
	if len(qr) < minFuzzyRunes {
		return 0, false
	}
	qi, gaps := 0, 0
	for i := 0; i < len(tr) && qi < len(qr); i++ {
		if tr[i] == qr[qi] {
			qi++
			continue
		}
		if qi > 0 {
			gaps++
		}
	}
	if qi < len(qr) || gaps > maxFuzzyGaps*len(qr) {
		return 0, false
	}
	bonus := 15 - gaps/4
	if bonus < 1 {
		bonus = 1
	}
	return bonus, true
}

// ScoreItem returns the relevance of query against it, or 0 when it does not
// match. An empty query matches everything with a neutral score, which keeps
// the default listing in catalog order.
//
// Every token must land somewhere, so multi-word queries narrow rather than
// widen: "secret aws" finds AWS secret rules, not everything mentioning either.
//
// Within a token the best title/group/summary hit is the base score and the
// best keyword hit is added at half weight. Keywords reinforce instead of
// competing, which is why "scan" ranks "Scan this project" (title + the "scan"
// keyword) above "Scan changes vs main" (title match alone).
func ScoreItem(query string, it Item) int {
	toks := tokenize(query)
	if len(toks) == 0 {
		return 1
	}
	fs := it.fields()
	total := 0
	for _, tok := range toks {
		best, keyword := 0, 0
		for _, f := range fs {
			s := scoreText(tok, f.text)
			if s == 0 {
				continue
			}
			s *= f.weight
			if f.keyword {
				if s > keyword {
					keyword = s
				}
			} else if s > best {
				best = s
			}
		}
		if best == 0 && keyword == 0 {
			return 0
		}
		total += best + keyword/2
	}
	return total
}

// Rank returns the items matching query, best first. Ties keep catalog order
// so the empty-query listing is stable across redraws.
func Rank(query string, items []Item) []Item {
	type scored struct {
		it    Item
		score int
		idx   int
	}
	hits := make([]scored, 0, len(items))
	for i, it := range items {
		if s := ScoreItem(query, it); s > 0 {
			hits = append(hits, scored{it, s, i})
		}
	}
	sort.SliceStable(hits, func(a, b int) bool {
		if hits[a].score != hits[b].score {
			return hits[a].score > hits[b].score
		}
		return hits[a].idx < hits[b].idx
	})
	out := make([]Item, len(hits))
	for i, h := range hits {
		out[i] = h.it
	}
	return out
}
