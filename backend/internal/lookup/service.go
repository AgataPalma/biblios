package lookup

import (
	"context"
	"fmt"
	"log/slog"
)

type Service struct {
	google      *googleBooksClient
	openLibrary *openLibraryClient
}

func NewService(googleAPIKey string) *Service {
	return &Service{
		google:      newGoogleBooksClient(googleAPIKey),
		openLibrary: newOpenLibraryClient(),
	}
}

func (s *Service) LookupByISBN(ctx context.Context, isbn string) (*GoogleBooksResult, error) {
	// Try Google Books first
	var result *GoogleBooksResult
	var err error

	result, err = s.google.SearchByISBN(ctx, isbn)
	if err != nil {
		slog.Warn("Google Books ISBN lookup failed, trying OpenLibrary", "error", err)
	}
	if result != nil {
		return result, nil
	}

	// Fallback to OpenLibrary
	result, err = s.openLibrary.SearchByISBN(ctx, isbn)
	if err != nil {
		return nil, fmt.Errorf("both lookup services failed: %w", err)
	}

	return result, nil
}

func (s *Service) LookupByTitleAuthor(ctx context.Context, title string, author string, page int) (*SearchResultList, error) {
	type apiResult struct {
		list *SearchResultList
		err  error
	}

	// ── Round 1: Portuguese only ─────────────────────────────────────────────
	ptGoogleCh := make(chan apiResult, 1)
	ptOlCh := make(chan apiResult, 1)

	go func() {
		r, err := s.google.SearchByTitleAuthor(ctx, title, author, page, "pt")
		ptGoogleCh <- apiResult{r, err}
	}()
	go func() {
		r, err := s.openLibrary.SearchByTitleAuthor(ctx, title, author, page, "por")
		ptOlCh <- apiResult{r, err}
	}()

	ptGoogle := <-ptGoogleCh
	ptOl := <-ptOlCh

	var ptResults []GoogleBooksResult
	if ptGoogle.list != nil {
		ptResults = append(ptResults, ptGoogle.list.Results...)
	}
	if ptOl.list != nil {
		ptResults = append(ptResults, ptOl.list.Results...)
	}

	// ── Round 2: Unrestricted ─────────────────────────────────────────────────
	googleCh := make(chan apiResult, 1)
	olCh := make(chan apiResult, 1)

	go func() {
		r, err := s.google.SearchByTitleAuthor(ctx, title, author, page, "")
		googleCh <- apiResult{r, err}
	}()
	go func() {
		r, err := s.openLibrary.SearchByTitleAuthor(ctx, title, author, page, "")
		olCh <- apiResult{r, err}
	}()

	googleRes := <-googleCh
	olRes := <-olCh

	if googleRes.err != nil {
		slog.Warn("Google Books lookup failed", "error", googleRes.err)
	}
	if olRes.err != nil {
		slog.Warn("OpenLibrary lookup failed", "error", olRes.err)
	}

	var allResults []GoogleBooksResult
	if googleRes.list != nil {
		allResults = append(allResults, googleRes.list.Results...)
	}
	if olRes.list != nil {
		allResults = append(allResults, olRes.list.Results...)
	}

	if len(ptResults) == 0 && len(allResults) == 0 {
		return nil, nil
	}

	// ── Merge: PT first, then unrestricted ────────────────────────────────────
	// Dedup key is isbn13+language so PT and EN editions of the same book
	// are kept as separate entries
	type dedupKey struct {
		isbn     string
		language string
	}
	seen := make(map[dedupKey]int)
	seenNoISBN := make(map[string]bool)
	var combined []GoogleBooksResult

	addResult := func(r GoogleBooksResult) {
		if r.ISBN13 != "" {
			k := dedupKey{r.ISBN13, r.Language}
			if idx, exists := seen[k]; exists {
				if metadataScore(r) > metadataScore(combined[idx]) {
					combined[idx] = r
				}
			} else {
				seen[k] = len(combined)
				combined = append(combined, r)
			}
		} else {
			k := r.Title + "|" + r.Language
			if !seenNoISBN[k] {
				seenNoISBN[k] = true
				combined = append(combined, r)
			}
		}
	}

	for _, r := range ptResults {
		addResult(r)
	}
	for _, r := range allResults {
		addResult(r)
	}

	// Use unrestricted total as the meaningful count
	var total int
	if googleRes.list != nil {
		total += googleRes.list.Total
	}
	if olRes.list != nil {
		total += olRes.list.Total
	}

	return &SearchResultList{
		Results:  combined,
		Total:    total,
		Page:     page,
		PageSize: 20,
	}, nil
}

// metadataScore counts how many fields are populated — higher = richer entry
func metadataScore(r GoogleBooksResult) int {
	var score int
	if r.Title != "" {
		score++
	}
	if len(r.Authors) > 0 {
		score++
	}
	if r.Publisher != "" {
		score++
	}
	if r.PublishedDate != "" {
		score++
	}
	if r.Description != "" {
		score++
	}
	if r.CoverURL != "" {
		score++
	}
	if r.PageCount > 0 {
		score++
	}
	if r.Language != "" {
		score++
	}
	if r.ISBN13 != "" {
		score++
	}
	if r.ISBN10 != "" {
		score++
	}
	return score
}
