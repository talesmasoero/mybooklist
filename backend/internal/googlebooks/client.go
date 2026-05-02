package googlebooks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/talesmasoero/mybooklist/backend/internal/domain"
)

const baseURL = "https://www.googleapis.com/books/v1/volumes"

type Client struct {
	httpClient *http.Client
	apiKey     string
}

func NewClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		apiKey:     apiKey,
	}
}

type volumesResponse struct {
	Items []volumeItem `json:"items"`
}

type volumeItem struct {
	ID         string     `json:"id"`
	VolumeInfo volumeInfo `json:"volumeInfo"`
}

type volumeInfo struct {
	Title               string              `json:"title"`
	Authors             []string            `json:"authors"`
	Categories          []string            `json:"categories"`
	Description         string              `json:"description"`
	PageCount           int                 `json:"pageCount"`
	ImageLinks          *imageLinks         `json:"imageLinks"`
	IndustryIdentifiers []industryIdentifier `json:"industryIdentifiers"`
}

type imageLinks struct {
	Thumbnail      string `json:"thumbnail"`
	SmallThumbnail string `json:"smallThumbnail"`
}

type industryIdentifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

func (c *Client) Search(ctx context.Context, query string, maxResults int) ([]domain.BookSearchResult, error) {
	if maxResults <= 0 || maxResults > 40 {
		maxResults = 10
	}

	u, _ := url.Parse(baseURL)
	q := u.Query()
	q.Set("q", query)
	q.Set("maxResults", strconv.Itoa(maxResults))
	if c.apiKey != "" {
		q.Set("key", c.apiKey)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: build request: %v", domain.ErrExternalAPIFailed, err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.WarnContext(ctx, "google books request failed", "error", err)
		return nil, fmt.Errorf("%w: %v", domain.ErrExternalAPIFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		slog.WarnContext(ctx, "google books non-2xx response", "status", resp.StatusCode)
		return nil, fmt.Errorf("%w: status %d", domain.ErrExternalAPIFailed, resp.StatusCode)
	}

	var body volumesResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		slog.WarnContext(ctx, "google books decode failed", "error", err)
		return nil, fmt.Errorf("%w: decode: %v", domain.ErrExternalAPIFailed, err)
	}

	results := make([]domain.BookSearchResult, 0, len(body.Items))
	for _, item := range body.Items {
		results = append(results, mapVolume(item))
	}
	return results, nil
}

func mapVolume(item volumeItem) domain.BookSearchResult {
	r := domain.BookSearchResult{
		GoogleBooksID: item.ID,
		Title:         item.VolumeInfo.Title,
		Authors:       item.VolumeInfo.Authors,
		Genres:        item.VolumeInfo.Categories,
		Synopsis:      item.VolumeInfo.Description,
		TotalPages:    item.VolumeInfo.PageCount,
	}
	if r.Authors == nil {
		r.Authors = []string{}
	}
	if r.Genres == nil {
		r.Genres = []string{}
	}
	if item.VolumeInfo.ImageLinks != nil {
		if item.VolumeInfo.ImageLinks.Thumbnail != "" {
			r.CoverURL = item.VolumeInfo.ImageLinks.Thumbnail
		} else {
			r.CoverURL = item.VolumeInfo.ImageLinks.SmallThumbnail
		}
	}
	for _, id := range item.VolumeInfo.IndustryIdentifiers {
		if id.Type == "ISBN_13" {
			r.ISBN = id.Identifier
			break
		}
		if id.Type == "ISBN_10" && r.ISBN == "" {
			r.ISBN = id.Identifier
		}
	}
	return r
}
