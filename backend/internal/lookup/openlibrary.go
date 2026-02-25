package lookup

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type openLibraryClient struct {
	httpClient *http.Client
}

type openLibraryISBNResponse struct {
	Title      string `json:"title"`
	Publishers []struct {
		Name string `json:"name"`
	} `json:"publishers"`
	PublishDate   string `json:"publish_date"`
	NumberOfPages int    `json:"number_of_pages"`
	Languages     []struct {
		Key string `json:"key"`
	} `json:"languages"`
	Authors []struct {
		Key string `json:"key"`
	} `json:"authors"`
	Cover struct {
		Large string `json:"large"`
	} `json:"cover"`
	Identifiers struct {
		ISBN13 []string `json:"isbn_13"`
		ISBN10 []string `json:"isbn_10"`
	} `json:"identifiers"`
}

type openLibrarySearchResponse struct {
	Docs []struct {
		Title               string   `json:"title"`
		AuthorNames         []string `json:"author_name"`
		Publisher           []string `json:"publisher"`
		FirstPublishYear    int      `json:"first_publish_year"`
		ISBN                []string `json:"isbn"`
		Language            []string `json:"language"`
		Subject             []string `json:"subject"`
		CoverI              int      `json:"cover_i"`
		NumberOfPagesMedian int      `json:"number_of_pages_median"`
	} `json:"docs"`
}

func newOpenLibraryClient() *openLibraryClient {
	return &openLibraryClient{
		httpClient: &http.Client{},
	}
}

func (c *openLibraryClient) SearchByISBN(ctx context.Context, isbn string) (*GoogleBooksResult, error) {
	var reqURL string = fmt.Sprintf("https://openlibrary.org/api/books?bibkeys=ISBN:%s&format=json&jscmd=data", isbn)

	var req *http.Request
	var err error
	req, err = http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	var resp *http.Response
	resp, err = c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call OpenLibrary API: %w", err)
	}
	defer resp.Body.Close()

	var raw map[string]openLibraryISBNResponse
	err = json.NewDecoder(resp.Body).Decode(&raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var key string = fmt.Sprintf("ISBN:%s", isbn)
	var book openLibraryISBNResponse
	var ok bool
	book, ok = raw[key]
	if !ok {
		return nil, nil
	}

	var result GoogleBooksResult = GoogleBooksResult{
		Title:         book.Title,
		PublishedDate: book.PublishDate,
		PageCount:     book.NumberOfPages,
		CoverURL:      book.Cover.Large,
	}

	if len(book.Publishers) > 0 {
		result.Publisher = book.Publishers[0].Name
	}
	if len(book.Identifiers.ISBN13) > 0 {
		result.ISBN13 = book.Identifiers.ISBN13[0]
	}
	if len(book.Identifiers.ISBN10) > 0 {
		result.ISBN10 = book.Identifiers.ISBN10[0]
	}

	return &result, nil
}

func (c *openLibraryClient) SearchByTitleAuthor(ctx context.Context, title string, author string) (*GoogleBooksResult, error) {
	var reqURL string = fmt.Sprintf(
		"https://openlibrary.org/search.json?title=%s&author=%s&limit=1",
		url.QueryEscape(title), url.QueryEscape(author),
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
		return nil, fmt.Errorf("failed to call OpenLibrary search: %w", err)
	}
	defer resp.Body.Close()

	var searchResp openLibrarySearchResponse
	err = json.NewDecoder(resp.Body).Decode(&searchResp)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(searchResp.Docs) == 0 {
		return nil, nil
	}

	var doc = searchResp.Docs[0]
	var result GoogleBooksResult = GoogleBooksResult{
		Title:      doc.Title,
		Authors:    doc.AuthorNames,
		Categories: doc.Subject,
		PageCount:  doc.NumberOfPagesMedian,
		Language:   strings.Join(doc.Language, ", "),
	}

	if len(doc.Publisher) > 0 {
		result.Publisher = doc.Publisher[0]
	}
	if len(doc.ISBN) > 0 {
		result.ISBN13 = doc.ISBN[0]
	}
	if doc.CoverI > 0 {
		result.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", doc.CoverI)
	}

	return &result, nil
}
