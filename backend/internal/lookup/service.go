package lookup

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type LookupResult struct {
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	ISBN10        string   `json:"isbn_10,omitempty"`
	ISBN13        string   `json:"isbn_13,omitempty"`
	Publisher     string   `json:"publisher,omitempty"`
	PublishedDate string   `json:"published_date,omitempty"`
	PageCount     int      `json:"page_count,omitempty"`
	CoverURL      string   `json:"cover_url,omitempty"`
	Language      string   `json:"language,omitempty"`
	Description   string   `json:"description,omitempty"`
	Categories    []string `json:"categories,omitempty"`
}

type Service struct {
	googleAPIKey string
	httpClient   *http.Client
}

func NewService(googleAPIKey string) *Service {
	return &Service{
		googleAPIKey: googleAPIKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// LookupByISBN performs a two-round concurrent lookup:
//
//	Round 1 – Google Books restricted to Portuguese (langRestrict=pt) + Open Library
//	Round 2 – Google Books unrestricted
//
// Results are merged with Portuguese entries prioritised and deduplicated by ISBN13+language.
func (s *Service) LookupByISBN(ctx context.Context, isbnRaw string) ([]LookupResult, error) {
	isbn := cleanISBN(isbnRaw)
	if isbn == "" {
		return nil, fmt.Errorf("isbn is required")
	}

	type roundResult struct {
		results []LookupResult
		err     error
	}

	// ── Round 1: Portuguese Google Books + Open Library (concurrent) ──────────
	r1Google := make(chan roundResult, 1)
	r1OL := make(chan roundResult, 1)

	go func() {
		res, err := s.googleBooksLookup(ctx, isbn, "pt")
		r1Google <- roundResult{res, err}
	}()
	go func() {
		res, err := s.openLibraryLookup(ctx, isbn)
		r1OL <- roundResult{res, err}
	}()

	g1 := <-r1Google
	ol := <-r1OL

	// ── Round 2: Google Books unrestricted ────────────────────────────────────
	r2Google := make(chan roundResult, 1)
	go func() {
		res, err := s.googleBooksLookup(ctx, isbn, "")
		r2Google <- roundResult{res, err}
	}()
	g2 := <-r2Google

	// ── Merge: Portuguese first, then unrestricted, then Open Library ─────────
	var merged []LookupResult
	seen := map[string]bool{} // key: isbn13+"|"+language

	addAll := func(results []LookupResult) {
		for _, r := range results {
			key := r.ISBN13 + "|" + strings.ToLower(r.Language)
			if key == "|" {
				// No ISBN13 — use title+language as dedup key
				key = strings.ToLower(r.Title) + "|" + strings.ToLower(r.Language)
			}
			if !seen[key] {
				seen[key] = true
				merged = append(merged, r)
			}
		}
	}

	// Portuguese results first (priority)
	if g1.err == nil {
		addAll(filterByLanguage(g1.results, "pt", "por"))
	}
	// Open Library results (often have Portuguese entries)
	if ol.err == nil {
		addAll(filterByLanguage(ol.results, "pt", "por"))
	}
	// Remaining unrestricted Google results
	if g2.err == nil {
		addAll(g2.results)
	}
	// Remaining Open Library results not yet added
	if ol.err == nil {
		addAll(ol.results)
	}

	return merged, nil
}

// GetCoverURL returns just the cover URL for a given ISBN (used by backfill).
func (s *Service) GetCoverURL(ctx context.Context, isbnRaw string) (string, error) {
	isbn := cleanISBN(isbnRaw)
	if isbn == "" {
		return "", nil
	}

	// Try Open Library cover first (no API key needed)
	olURL := fmt.Sprintf("https://covers.openlibrary.org/b/isbn/%s-L.jpg?default=false", isbn)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, olURL, nil)
	if err == nil {
		resp, err := s.httpClient.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return olURL, nil
		}
		if resp != nil {
			resp.Body.Close()
		}
	}

	// Fall back to Google Books thumbnail
	results, err := s.googleBooksLookup(ctx, isbn, "")
	if err != nil || len(results) == 0 {
		return "", nil
	}
	return results[0].CoverURL, nil
}

// ─── Google Books ─────────────────────────────────────────────────────────────

type googleBooksResponse struct {
	Items []struct {
		VolumeInfo struct {
			Title               string   `json:"title"`
			Authors             []string `json:"authors"`
			Publisher           string   `json:"publisher"`
			PublishedDate       string   `json:"publishedDate"`
			Description         string   `json:"description"`
			PageCount           int      `json:"pageCount"`
			Language            string   `json:"language"`
			Categories          []string `json:"categories"`
			IndustryIdentifiers []struct {
				Type       string `json:"type"`
				Identifier string `json:"identifier"`
			} `json:"industryIdentifiers"`
			ImageLinks struct {
				Thumbnail string `json:"thumbnail"`
			} `json:"imageLinks"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

func (s *Service) googleBooksLookup(ctx context.Context, isbn, langRestrict string) ([]LookupResult, error) {
	params := url.Values{}
	params.Set("q", "isbn:"+isbn)
	params.Set("maxResults", "10")
	if langRestrict != "" {
		params.Set("langRestrict", langRestrict)
	}
	if s.googleAPIKey != "" {
		params.Set("key", s.googleAPIKey)
	}

	reqURL := "https://www.googleapis.com/books/v1/volumes?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build google books request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google books request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google books returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read google books response: %w", err)
	}

	var gbResp googleBooksResponse
	if err := json.Unmarshal(body, &gbResp); err != nil {
		return nil, fmt.Errorf("parse google books response: %w", err)
	}

	var results []LookupResult
	for _, item := range gbResp.Items {
		vi := item.VolumeInfo
		r := LookupResult{
			Title:         vi.Title,
			Authors:       vi.Authors,
			Publisher:     vi.Publisher,
			PublishedDate: vi.PublishedDate,
			PageCount:     vi.PageCount,
			Language:      vi.Language,
			Description:   vi.Description,
			Categories:    vi.Categories,
		}
		// Extract cover URL — prefer HTTPS
		if vi.ImageLinks.Thumbnail != "" {
			r.CoverURL = strings.Replace(vi.ImageLinks.Thumbnail, "http://", "https://", 1)
		}
		// Extract ISBNs
		for _, id := range vi.IndustryIdentifiers {
			switch id.Type {
			case "ISBN_10":
				r.ISBN10 = id.Identifier
			case "ISBN_13":
				r.ISBN13 = id.Identifier
			}
		}
		results = append(results, r)
	}
	return results, nil
}

// ─── Open Library ─────────────────────────────────────────────────────────────

type openLibraryResponse map[string]struct {
	Title   string `json:"title"`
	Authors []struct {
		Name string `json:"name"`
	} `json:"authors"`
	Publishers []struct {
		Name string `json:"name"`
	} `json:"publishers"`
	PublishDate   string `json:"publish_date"`
	NumberOfPages int    `json:"number_of_pages"`
	Languages     []struct {
		Key string `json:"key"`
	} `json:"languages"`
	Subjects []string `json:"subjects"`
	Cover    struct {
		Large string `json:"large"`
	} `json:"cover"`
	ISBNs10 []string `json:"isbn_10"`
	ISBNs13 []string `json:"isbn_13"`
}

func (s *Service) openLibraryLookup(ctx context.Context, isbn string) ([]LookupResult, error) {
	reqURL := fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", isbn)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build open library request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("open library request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("open library returned %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read open library response: %w", err)
	}

	var olResp openLibraryResponse
	if err := json.Unmarshal(body, &olResp); err != nil {
		return nil, fmt.Errorf("parse open library response: %w", err)
	}

	var results []LookupResult
	for _, book := range olResp {
		r := LookupResult{
			Title:         book.Title,
			PublishedDate: book.PublishDate,
			PageCount:     book.NumberOfPages,
		}
		for _, a := range book.Authors {
			r.Authors = append(r.Authors, a.Name)
		}
		if len(book.Publishers) > 0 {
			r.Publisher = book.Publishers[0].Name
		}
		if len(book.Languages) > 0 {
			// Open Library language keys look like "/languages/por"
			parts := strings.Split(book.Languages[0].Key, "/")
			r.Language = parts[len(parts)-1]
		}
		r.Categories = book.Subjects
		if book.Cover.Large != "" {
			r.CoverURL = book.Cover.Large
		}
		if len(book.ISBNs10) > 0 {
			r.ISBN10 = book.ISBNs10[0]
		}
		if len(book.ISBNs13) > 0 {
			r.ISBN13 = book.ISBNs13[0]
		}
		results = append(results, r)
	}
	return results, nil
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func cleanISBN(s string) string {
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, " ", "")
	return strings.TrimSpace(s)
}

func filterByLanguage(results []LookupResult, langs ...string) []LookupResult {
	set := map[string]bool{}
	for _, l := range langs {
		set[strings.ToLower(l)] = true
	}
	var out []LookupResult
	for _, r := range results {
		if set[strings.ToLower(r.Language)] {
			out = append(out, r)
		}
	}
	return out
}
