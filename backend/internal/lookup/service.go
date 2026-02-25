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

func (s *Service) LookupByTitleAuthor(ctx context.Context, title string, author string) (*GoogleBooksResult, error) {
	// Try Google Books first
	var result *GoogleBooksResult
	var err error

	result, err = s.google.SearchByTitleAuthor(ctx, title, author)
	if err != nil {
		slog.Warn("Google Books title/author lookup failed, trying OpenLibrary", "error", err)
	}
	if result != nil {
		return result, nil
	}

	// Fallback to OpenLibrary
	result, err = s.openLibrary.SearchByTitleAuthor(ctx, title, author)
	if err != nil {
		return nil, fmt.Errorf("both lookup services failed: %w", err)
	}

	return result, nil
}
