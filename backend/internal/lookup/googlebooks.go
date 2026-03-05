package lookup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type googleBooksClient struct {
	apiKey     string
	httpClient *http.Client
}

type GoogleBooksResult struct {
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	Publisher     string   `json:"publisher"`
	PublishedDate string   `json:"published_date"`
	Description   string   `json:"description"`
	ISBN10        string   `json:"isbn_10"`
	ISBN13        string   `json:"isbn_13"`
	PageCount     int      `json:"page_count"`
	Language      string   `json:"language"`
	CoverURL      string   `json:"cover_url"`
	Categories    []string `json:"categories"`
}

type googleBooksResponse struct {
	TotalItems int `json:"totalItems"`
	Items      []struct {
		VolumeInfo struct {
			Title         string   `json:"title"`
			Authors       []string `json:"authors"`
			Publisher     string   `json:"publisher"`
			PublishedDate string   `json:"publishedDate"`
			Description   string   `json:"description"`
			PageCount     int      `json:"pageCount"`
			Language      string   `json:"language"`
			Categories    []string `json:"categories"`
			ImageLinks    struct {
				Thumbnail string `json:"thumbnail"`
			} `json:"imageLinks"`
			IndustryIdentifiers []struct {
				Type       string `json:"type"`
				Identifier string `json:"identifier"`
			} `json:"industryIdentifiers"`
		} `json:"volumeInfo"`
	} `json:"items"`
}

type SearchResultList struct {
	Results  []GoogleBooksResult `json:"results"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

func newGoogleBooksClient(apiKey string) *googleBooksClient {
	return &googleBooksClient{
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

func (c *googleBooksClient) SearchByISBN(ctx context.Context, isbn string) (*GoogleBooksResult, error) {
	var query string = fmt.Sprintf("isbn:%s", isbn)
	return c.search(ctx, query)
}

// Keep SearchByISBN unchanged — it calls the existing search() method

// Replace SearchByTitleAuthor:
func (c *googleBooksClient) SearchByTitleAuthor(ctx context.Context, title string, author string, page int, lang string) (*SearchResultList, error) {
	var pageSize int = 20
	var startIndex int = (page - 1) * pageSize
	var query string = fmt.Sprintf("intitle:%s+inauthor:%s", url.QueryEscape(title), url.QueryEscape(author))
	var reqURL string = fmt.Sprintf(
		"https://www.googleapis.com/books/v1/volumes?q=%s&key=%s&maxResults=%d&startIndex=%d",
		query, c.apiKey, pageSize, startIndex,
	)
	if lang != "" {
		reqURL += "&langRestrict=" + lang
	}

	var req *http.Request
	var err error
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp *http.Response
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Books API: %w", err)
	}
	defer resp.Body.Close()

	var gbResp googleBooksResponse
	err = json.NewDecoder(resp.Body).Decode(&gbResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var results []GoogleBooksResult
	for _, item := range gbResp.Items {
		var v = item.VolumeInfo
		var r GoogleBooksResult = GoogleBooksResult{
			Title:         v.Title,
			Authors:       v.Authors,
			Publisher:     v.Publisher,
			PublishedDate: v.PublishedDate,
			Description:   v.Description,
			PageCount:     v.PageCount,
			Language:      v.Language,
			CoverURL:      v.ImageLinks.Thumbnail,
			Categories:    v.Categories,
		}
		for _, id := range v.IndustryIdentifiers {
			if id.Type == "ISBN_13" {
				r.ISBN13 = id.Identifier
			}
			if id.Type == "ISBN_10" {
				r.ISBN10 = id.Identifier
			}
		}
		results = append(results, r)
	}

	return &SearchResultList{
		Results:  results,
		Total:    gbResp.TotalItems,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (c *googleBooksClient) search(ctx context.Context, query string) (*GoogleBooksResult, error) {
	var reqURL string = fmt.Sprintf(
		"https://www.googleapis.com/books/v1/volumes?q=%s&key=%s",
		query, c.apiKey,
	)

	var req *http.Request
	var err error
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp *http.Response
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Google Books API: %w", err)
	}
	defer resp.Body.Close()

	var gbResp googleBooksResponse
	err = json.NewDecoder(resp.Body).Decode(&gbResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(gbResp.Items) == 0 {
		return nil, nil
	}

	var item = gbResp.Items[0].VolumeInfo
	var result GoogleBooksResult = GoogleBooksResult{
		Title:         item.Title,
		Authors:       item.Authors,
		Publisher:     item.Publisher,
		PublishedDate: item.PublishedDate,
		Description:   item.Description,
		PageCount:     item.PageCount,
		Language:      item.Language,
		CoverURL:      item.ImageLinks.Thumbnail,
		Categories:    item.Categories,
	}

	for _, id := range item.IndustryIdentifiers {
		if id.Type == "ISBN_13" {
			result.ISBN13 = id.Identifier
		}
		if id.Type == "ISBN_10" {
			result.ISBN10 = id.Identifier
		}
	}

	return &result, nil
}
