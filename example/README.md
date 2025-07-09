# Jackett Go Client Example

This directory contains an example application demonstrating how to use the Jackett Go client library.

## Prerequisites

1. A running Jackett instance
2. Your Jackett API key
3. Go 1.23 or later

## Setup

1. Edit `main.go` and replace `"your-api-key"` with your actual Jackett API key
2. Update the address and port if your Jackett instance is not running on `localhost:9117`

## Running the Example

```bash
# Build the example
go build -o jackett-example .

# Run the example
./jackett-example
```

## What the Example Does

The example demonstrates:

1. **Server Configuration**: Retrieves and displays Jackett server information
2. **Indexer Management**: Lists all configured indexers
3. **Search Functionality**: Performs a sample search for "The Matrix 1999"

## Expected Output

If everything is configured correctly, you should see output similar to:

```
✓ Successfully created Jackett client

Getting server configuration...
✓ Jackett version: 0.20.1234.0
✓ API port: 9117

Getting configured indexers...
✓ Found 3 configured indexers
  - RARBG (public)
  - The Pirate Bay (public)
  - 1337x (public)

Searching for 'The Matrix 1999'...
✓ Found 15 results
  1. The Matrix 1999 1080p BluRay x264 (45 seeders, 1.2 GB)
  2. The Matrix 1999 2160p UHD BluRay x265 (23 seeders, 4.5 GB)
  3. The Matrix 1999 720p BluRay x264 (67 seeders, 800.5 MB)
... and 12 more results

Example completed successfully!
```

## Troubleshooting

- **Connection failed**: Make sure Jackett is running and accessible
- **Invalid API key**: Verify your API key in Jackett's web interface
- **No indexers found**: Configure at least one indexer in Jackett
- **Search returns no results**: Some indexers may be down or have no results for the search term 