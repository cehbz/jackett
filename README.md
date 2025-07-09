# Jackett Go Client Library

A Go client library for interacting with the [Jackett](https://github.com/Jackett/Jackett) API.

## Features

- **Search**: Search across all configured indexers or specific indexers
- **Indexer Management**: Retrieve information about configured indexers
- **Torrent Download**: Download torrent files from search results
- **Server Configuration**: Access Jackett server configuration

## Installation

To install the package, run:

```bash
go get github.com/cehbz/jackett
```

## Usage

### Importing the Package

```go
import (
    "github.com/cehbz/jackett"
)
```

### Initializing the Client

```go
client, err := jackett.NewClient("http://localhost:9117", "your-api-key")
if err != nil {
    log.Fatalf("Failed to create client: %v", err)
}
```

- `baseURL`: The full URL where Jackett is running (e.g., "http://127.0.0.1:9117")
- `apiKey`: Your Jackett API key

### Searching for Torrents

#### Search All Indexers
```go
results, err := client.Search("The Matrix 1999")
if err != nil {
    log.Fatalf("Search failed: %v", err)
}

for _, result := range results.Results {
    fmt.Printf("Title: %s\n", result.Title)
    fmt.Printf("Size: %d bytes\n", result.Size)
    fmt.Printf("Seeders: %d\n", result.Seeders)
    fmt.Printf("Tracker: %s\n", result.Tracker)
}
```

#### Search Specific Indexer
```go
results, err := client.SearchWithIndexer("rarbg", "The Matrix 1999")
if err != nil {
    log.Fatalf("Search failed: %v", err)
}
```

### Managing Indexers

```go
indexers, err := client.GetIndexers()
if err != nil {
    log.Fatalf("Failed to get indexers: %v", err)
}

for _, indexer := range indexers {
    fmt.Printf("Name: %s\n", indexer.Name)
    fmt.Printf("Type: %s\n", indexer.Type)
    fmt.Printf("Configured: %v\n", indexer.Configured)
}
```

### Downloading Torrents

```go
// Download torrent from search result
torrentData, err := client.DownloadTorrent(result.Link)
if err != nil {
    log.Fatalf("Failed to download torrent: %v", err)
}

// Save to file
err = os.WriteFile("movie.torrent", torrentData, 0644)
if err != nil {
    log.Fatalf("Failed to save torrent file: %v", err)
}
```

### Getting Server Configuration

```go
config, err := client.GetServerConfig()
if err != nil {
    log.Fatalf("Failed to get server config: %v", err)
}

fmt.Printf("Jackett version: %v\n", config["app_version"])
fmt.Printf("API port: %v\n", config["port"])
```

## Connection Handling

The client does not provide a separate connection test method. Connection and authentication are implicitly tested by attempting to fetch indexers, perform a search, or retrieve server configuration. If there is a problem, the relevant method will return an error.

## Search Result Categories

Jackett uses numeric category IDs. Common categories include:

- **Console**: 1000-1999
- **Movies**: 2000-2999
- **Audio**: 3000-3999
- **PC Games**: 4000-4999
- **TV**: 5000-5999
- **XXX**: 6000-6999
- **Books**: 7000-7999
- **Other**: 8000-8999

## Error Handling

The client returns detailed errors for various failure scenarios:

```go
results, err := client.Search("test")
if err != nil {
    // Handle specific error cases
    switch {
    case strings.Contains(err.Error(), "401"):
        log.Fatal("Invalid API key")
    case strings.Contains(err.Error(), "connection refused"):
        log.Fatal("Jackett is not running")
    default:
        log.Fatalf("Search error: %v", err)
    }
}
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contribution

Contributions are welcome! Please open an issue or submit a pull request for any improvements or bug fixes.

## Acknowledgments

- [Jackett API Documentation](https://github.com/Jackett/Jackett/wiki/Jackett-API)
