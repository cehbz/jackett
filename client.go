package jackett

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client is a Jackett API client. It is immutable and safe for concurrent use.
type Client struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

// SearchResult represents a torrent search result from Jackett
type SearchResult struct {
	Title                string    `json:"Title"`
	Size                 int64     `json:"Size"`
	Seeders              int       `json:"Seeders"`
	Peers                int       `json:"Peers"`
	Link                 string    `json:"Link"`
	MagnetURI            string    `json:"MagnetUri"`
	GUID                 string    `json:"Guid"`
	PublishDate          string    `json:"PublishDate"`
	Tracker              string    `json:"Tracker"`
	Category             []int     `json:"Category"`
	CategoryDesc         string    `json:"CategoryDesc"`
	BlackholeLink        *string   `json:"BlackholeLink"`
	Gain                 float64   `json:"Gain"`
	InfoHash             string    `json:"InfoHash"`
	MinimumRatio         *float64  `json:"MinimumRatio,omitempty"`
	MinimumSeedTime      *int64    `json:"MinimumSeedTime,omitempty"`
	DownloadVolumeFactor float64   `json:"DownloadVolumeFactor"`
	UploadVolumeFactor   float64   `json:"UploadVolumeFactor"`
	FirstSeen            string    `json:"FirstSeen"`
	TrackerId            string    `json:"TrackerId"`
	TrackerType          string    `json:"TrackerType"`
	Details              string    `json:"Details"`
	Files                *int      `json:"Files"`
	Grabs                *int      `json:"Grabs"`
	Description          *string   `json:"Description"`
	RageID               *int      `json:"RageID"`
	TVDBId               *int      `json:"TVDBId"`
	Imdb                 *int      `json:"Imdb"`
	TMDb                 *int      `json:"TMDb"`
	TVMazeId             *int      `json:"TVMazeId"`
	TraktId              *int      `json:"TraktId"`
	DoubanId             *int      `json:"DoubanId"`
	Genres               *[]string `json:"Genres"`
	Languages            []string  `json:"Languages"`
	Subs                 []string  `json:"Subs"`
	Year                 *int      `json:"Year"`
	Author               *string   `json:"Author"`
	BookTitle            *string   `json:"BookTitle"`
	Publisher            *string   `json:"Publisher"`
	Artist               *string   `json:"Artist"`
	Album                *string   `json:"Album"`
	Label                *string   `json:"Label"`
	Track                *string   `json:"Track"`
	Poster               *string   `json:"Poster"`
}

// SearchResponse represents the response from a search query
type SearchResponse struct {
	Results  []SearchResult `json:"Results"`
	Indexers []struct {
		ID      string `json:"ID"`
		Name    string `json:"Name"`
		Status  int    `json:"Status"`
		Results int64  `json:"Results"`
		Error   string `json:"Error"`
	} `json:"Indexers"`
}

// Indexer represents a configured indexer in Jackett
type Indexer struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Type        string     `json:"type"`
	Configured  bool       `json:"configured"`
	SiteLink    string     `json:"site_link"`
	Language    string     `json:"language"`
	Caps        *Caps      `json:"caps,omitempty"`
	Categories  []Category `json:"categories,omitempty"`
}

type Caps struct {
	Server    string    `json:"server"`
	Limits    Limits    `json:"limits"`
	Searching Searching `json:"searching"`
}

type Limits struct {
	Default string `json:"default"`
	Max     string `json:"max"`
}

type Searching struct {
	Search      *SearchType `json:"search,omitempty"`
	TVSearch    *SearchType `json:"tv_search,omitempty"`
	MovieSearch *SearchType `json:"movie_search,omitempty"`
	MusicSearch *SearchType `json:"music_search,omitempty"`
	AudioSearch *SearchType `json:"audio_search,omitempty"`
	BookSearch  *SearchType `json:"book_search,omitempty"`
}

type SearchType struct {
	Available       string `json:"available"`
	SupportedParams string `json:"supported_params"`
}

type Category struct {
	ID      int      `json:"id"`
	Name    string   `json:"name"`
	Subcats []Subcat `json:"subcats,omitempty"`
}

type Subcat struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// TorznabIndexersResponse represents the XML response from the torznab indexers endpoint
// Updated to match the real Jackett XML structure, including all fields
type TorznabIndexersResponse struct {
	XMLName  xml.Name         `xml:"indexers"`
	Indexers []TorznabIndexer `xml:"indexer"`
}

type TorznabIndexer struct {
	ID          string      `xml:"id,attr"`
	Configured  bool        `xml:"configured,attr"`
	Title       string      `xml:"title"`
	Description string      `xml:"description"`
	Link        string      `xml:"link"`
	Language    string      `xml:"language"`
	Type        string      `xml:"type"`
	Caps        TorznabCaps `xml:"caps"`
}

type TorznabCaps struct {
	Server     TorznabServer     `xml:"server"`
	Limits     TorznabLimits     `xml:"limits"`
	Searching  TorznabSearching  `xml:"searching"`
	Categories TorznabCategories `xml:"categories"`
}

type TorznabServer struct {
	Title string `xml:"title,attr"`
}

type TorznabLimits struct {
	Default string `xml:"default,attr"`
	Max     string `xml:"max,attr"`
}

type TorznabSearching struct {
	Search      *TorznabSearchType `xml:"search"`
	TVSearch    *TorznabSearchType `xml:"tv-search"`
	MovieSearch *TorznabSearchType `xml:"movie-search"`
	MusicSearch *TorznabSearchType `xml:"music-search"`
	AudioSearch *TorznabSearchType `xml:"audio-search"`
	BookSearch  *TorznabSearchType `xml:"book-search"`
}

type TorznabSearchType struct {
	Available       string `xml:"available,attr"`
	SupportedParams string `xml:"supportedParams,attr"`
}

type TorznabCategories struct {
	Categories []TorznabCategory `xml:"category"`
}

type TorznabCategory struct {
	ID      int             `xml:"id,attr"`
	Name    string          `xml:"name,attr"`
	Subcats []TorznabSubcat `xml:"subcat"`
}

type TorznabSubcat struct {
	ID   int    `xml:"id,attr"`
	Name string `xml:"name,attr"`
}

// NewClient initializes a new Jackett client.
// baseURL should be the full URL to the Jackett instance, e.g. "http://localhost:9117"
// If httpClient is nil, http.DefaultClient is used.
func NewClient(baseURL, apiKey string, httpClient ...*http.Client) (*Client, error) {
	client := http.DefaultClient
	if len(httpClient) > 0 && httpClient[0] != nil {
		client = httpClient[0]
	}

	jClient := &Client{
		client:  client,
		baseURL: baseURL,
		apiKey:  apiKey,
	}

	return jClient, nil
}

// Search performs a search query across all configured indexers
func (c *Client) Search(query string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("Query", query)

	respData, err := c.doGet("/api/v2.0/indexers/all/results", params)
	if err != nil {
		return nil, fmt.Errorf("search error: %v", err)
	}

	var response SearchResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %v", err)
	}

	return &response, nil
}

// SearchWithIndexer performs a search query on a specific indexer
func (c *Client) SearchWithIndexer(indexerID, query string) (*SearchResponse, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("Query", query)

	endpoint := fmt.Sprintf("/api/v2.0/indexers/%s/results", indexerID)
	respData, err := c.doGet(endpoint, params)
	if err != nil {
		return nil, fmt.Errorf("search error: %v", err)
	}

	var response SearchResponse
	if err := json.Unmarshal(respData, &response); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %v", err)
	}

	return &response, nil
}

// GetIndexers retrieves all configured indexers
func (c *Client) GetIndexers() ([]Indexer, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)
	params.Set("t", "indexers")
	params.Set("configured", "true")

	respData, err := c.doGet("/api/v2.0/indexers/all/results/torznab", params)
	if err != nil {
		return nil, fmt.Errorf("get indexers error: %v", err)
	}

	var torznabResponse TorznabIndexersResponse
	if err := xml.Unmarshal(respData, &torznabResponse); err != nil {
		return nil, fmt.Errorf("failed to decode indexers response: %v", err)
	}

	// Convert TorznabIndexer to Indexer
	indexers := make([]Indexer, len(torznabResponse.Indexers))
	for i, tIdx := range torznabResponse.Indexers {
		// Convert caps
		caps := &Caps{
			Server: tIdx.Caps.Server.Title,
			Limits: Limits{
				Default: tIdx.Caps.Limits.Default,
				Max:     tIdx.Caps.Limits.Max,
			},
			Searching: Searching{
				Search:      convertSearchType(tIdx.Caps.Searching.Search),
				TVSearch:    convertSearchType(tIdx.Caps.Searching.TVSearch),
				MovieSearch: convertSearchType(tIdx.Caps.Searching.MovieSearch),
				MusicSearch: convertSearchType(tIdx.Caps.Searching.MusicSearch),
				AudioSearch: convertSearchType(tIdx.Caps.Searching.AudioSearch),
				BookSearch:  convertSearchType(tIdx.Caps.Searching.BookSearch),
			},
		}
		// Convert categories
		categories := make([]Category, len(tIdx.Caps.Categories.Categories))
		for j, cat := range tIdx.Caps.Categories.Categories {
			subcats := make([]Subcat, len(cat.Subcats))
			for k, sub := range cat.Subcats {
				subcats[k] = Subcat(sub)
			}
			categories[j] = Category{ID: cat.ID, Name: cat.Name, Subcats: subcats}
		}
		indexers[i] = Indexer{
			ID:          tIdx.ID,
			Name:        tIdx.Title,
			Description: tIdx.Description,
			Type:        tIdx.Type,
			Configured:  tIdx.Configured,
			SiteLink:    tIdx.Link,
			Language:    tIdx.Language,
			Caps:        caps,
			Categories:  categories,
		}
	}

	return indexers, nil
}

func convertSearchType(t *TorznabSearchType) *SearchType {
	if t == nil {
		return nil
	}
	return &SearchType{
		Available:       t.Available,
		SupportedParams: t.SupportedParams,
	}
}

// DownloadTorrent downloads a torrent file from the given link
func (c *Client) DownloadTorrent(link string) ([]byte, error) {
	// Parse the link to check if it's a Jackett URL
	linkURL, err := url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("invalid download link: %v", err)
	}

	// If it's not already pointing to this Jackett instance, use it as-is
	baseURL, _ := url.Parse(c.baseURL)
	if linkURL.Host != baseURL.Host {
		// External link, download directly
		resp, err := c.client.Get(link)
		if err != nil {
			return nil, fmt.Errorf("download error: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(body))
		}

		return io.ReadAll(resp.Body)
	}

	// It's a Jackett link, ensure API key is present
	query := linkURL.Query()
	if query.Get("apikey") == "" {
		query.Set("apikey", c.apiKey)
		linkURL.RawQuery = query.Encode()
	}

	resp, err := c.client.Get(linkURL.String())
	if err != nil {
		return nil, fmt.Errorf("download error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// doGet is a helper method for making GET requests to the Jackett API
func (c *Client) doGet(endpoint string, query url.Values) ([]byte, error) {
	apiURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %v", err)
	}

	apiURL.Path = endpoint
	apiURL.RawQuery = query.Encode()

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected response code: %d, response: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// GetServerConfig retrieves the Jackett server configuration
func (c *Client) GetServerConfig() (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("apikey", c.apiKey)

	respData, err := c.doGet("/api/v2.0/server/config", params)
	if err != nil {
		return nil, fmt.Errorf("get server config error: %v", err)
	}

	var config map[string]interface{}
	if err := json.Unmarshal(respData, &config); err != nil {
		return nil, fmt.Errorf("failed to decode server config: %v", err)
	}

	return config, nil
}
