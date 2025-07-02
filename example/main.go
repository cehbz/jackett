package main

import (
	"fmt"
	"log"

	"github.com/cehbz/jackett"
)

func main() {
	// Initialize the client
	// Replace with your actual Jackett base URL and API key
	client, err := jackett.NewClient("http://localhost:9117", "your-api-key")
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	// Test the connection
	fmt.Println("Testing connection to Jackett...")
	err = client.TestConnection()
	if err != nil {
		log.Fatalf("Connection test failed: %v", err)
	}
	fmt.Println("✓ Successfully connected to Jackett")

	// Get server configuration
	fmt.Println("\nGetting server configuration...")
	config, err := client.GetServerConfig()
	if err != nil {
		log.Printf("Warning: Failed to get server config: %v", err)
	} else {
		fmt.Printf("✓ Jackett version: %v\n", config["app_version"])
		fmt.Printf("✓ API port: %v\n", config["port"])
	}

	// Get configured indexers
	fmt.Println("\nGetting configured indexers...")
	indexers, err := client.GetIndexers()
	if err != nil {
		log.Fatalf("Failed to get indexers: %v", err)
	}
	fmt.Printf("✓ Found %d configured indexers\n", len(indexers))

	// List indexers
	for i, indexer := range indexers {
		if i >= 5 { // Limit to first 5 for brevity
			fmt.Printf("... and %d more\n", len(indexers)-5)
			break
		}
		fmt.Printf("  - %s (%s)\n", indexer.Name, indexer.Type)
	}

	// Perform a search (only if you have indexers configured)
	if len(indexers) > 0 {
		fmt.Println("\nSearching for 'The Matrix 1999'...")
		results, err := client.Search("The Matrix 1999")
		if err != nil {
			log.Printf("Warning: Search failed: %v", err)
		} else {
			fmt.Printf("✓ Found %d results\n", len(results.Results))

			// Show first few results
			for i, result := range results.Results {
				if i >= 3 { // Limit to first 3 results
					fmt.Printf("... and %d more results\n", len(results.Results)-3)
					break
				}
				fmt.Printf("  %d. %s (%d seeders, %s)\n",
					i+1,
					result.Title,
					result.Seeders,
					formatSize(result.Size))
			}
		}
	} else {
		fmt.Println("\nNo indexers configured. Please configure at least one indexer in Jackett.")
	}

	fmt.Println("\nExample completed successfully!")
}

// formatSize converts bytes to human readable format
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
