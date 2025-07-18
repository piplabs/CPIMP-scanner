package main

import (
	"bufio"
	"log"
	"os"
	"strings"
	"time"
)

// Example configurations for different scanning scenarios

// Scan Base network with default settings
func BaseNetworkConfig() ScannerConfig {
	return ScannerConfig{
		Network:    "base",
		EventTopic: "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", // Upgraded(address)
		BlockRange: 10000,
		RateLimit:  500 * time.Millisecond,
		StartBlock: 0,
		EndBlock:   0, // 0 means latest
		OutputFile: "base_upgraded_transactions.csv",
	}
}

// Scan Ethereum network with smaller block ranges (due to higher activity)
func EthereumNetworkConfig() ScannerConfig {
	return ScannerConfig{
		Network:    "ethereum",
		EventTopic: "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", // Upgraded(address)
		BlockRange: 5000,                                                                 // Smaller range for Ethereum
		RateLimit:  1000 * time.Millisecond,                                              // Slower rate limit
		StartBlock: 0,
		EndBlock:   0,
		OutputFile: "ethereum_upgraded_transactions.csv",
	}
}

// Scan only recent blocks (last 100,000 blocks)
func RecentBlocksConfig() ScannerConfig {
	return ScannerConfig{
		Network:    "base",
		EventTopic: "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b",
		BlockRange: 10000,
		RateLimit:  300 * time.Millisecond,
		StartBlock: 0, // Will be calculated as latest - 100000
		EndBlock:   0, // Latest
		OutputFile: "recent_upgraded_transactions.csv",
	}
}

// Fast scan with larger block ranges (use with caution)
func FastScanConfig() ScannerConfig {
	return ScannerConfig{
		Network:    "base",
		EventTopic: "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b",
		BlockRange: 50000,                  // Larger range
		RateLimit:  200 * time.Millisecond, // Faster rate
		StartBlock: 0,
		EndBlock:   0,
		OutputFile: "fast_scan_results.csv",
	}
}

// Custom event scanner (example for different event)
func CustomEventConfig() ScannerConfig {
	return ScannerConfig{
		Network:    "base",
		EventTopic: "0x...", // Replace with your event topic hash
		BlockRange: 10000,
		RateLimit:  500 * time.Millisecond,
		StartBlock: 0,
		EndBlock:   0,
		OutputFile: "custom_event_results.csv",
	}
}

// Scan Story network with default settings
func StoryNetworkConfig() ScannerConfig {
	return ScannerConfig{
		Network:         "story",
		EventTopic:      "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", // Upgraded(address)
		BlockRange:      10000,
		RateLimit:       500 * time.Millisecond,
		StartBlock:      0,
		EndBlock:        0, // 0 means latest
		OutputFile:      "story_upgraded_transactions.csv",
		TargetAddresses: []string{}, // Empty = scan all addresses
	}
}

// Scan specific addresses only (much faster)
func StoryTargetedScanConfig() ScannerConfig {
	return ScannerConfig{
		Network:    "story",
		EventTopic: "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", // Upgraded(address)
		BlockRange: 50000,                                                                // Larger range since we're filtering by address
		RateLimit:  300 * time.Millisecond,
		StartBlock: 0,
		EndBlock:   0,
		OutputFile: "story_targeted_scan.csv",
		TargetAddresses: []string{
			"0x1234567890123456789012345678901234567890", // Replace with actual addresses
			"0x2345678901234567890123456789012345678901",
			"0x3456789012345678901234567890123456789012",
		},
	}
}

// Load addresses from file configuration
func StoryAddressListConfig(addressFile string) ScannerConfig {
	addresses := loadAddressesFromFile(addressFile)
	return ScannerConfig{
		Network:         "story",
		EventTopic:      "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b",
		BlockRange:      50000,
		RateLimit:       300 * time.Millisecond,
		StartBlock:      0,
		EndBlock:        0,
		OutputFile:      "story_address_list_scan.csv",
		TargetAddresses: addresses,
	}
}

func EthereumNetworkListConfig(addressFile string) ScannerConfig {
	addresses := loadAddressesFromFile(addressFile)
	return ScannerConfig{
		Network:         "ethereum",
		EventTopic:      "0xbc7cd75a20ee27fd9adebab32041f755214dbc6bffa90cc0225b39da2e5c2d3b", // Upgraded(address)
		BlockRange:      500,                                                                  // Smaller range for Ethereum
		RateLimit:       1000 * time.Millisecond,                                              // Slower rate limit
		StartBlock:      0,
		EndBlock:        22830367 + 100,
		OutputFile:      "ethereum_address_list_scan.csv",
		TargetAddresses: addresses,
	}
}

// Helper function to load addresses from a file
func loadAddressesFromFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		log.Printf("Warning: Could not open address file %s: %v", filename, err)
		return []string{}
	}
	defer file.Close()

	var addresses []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		address := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if address != "" && !strings.HasPrefix(address, "#") && !strings.HasPrefix(address, "//") {
			addresses = append(addresses, address)
		}
	}

	if err := scanner.Err(); err != nil {
		log.Printf("Error reading address file %s: %v", filename, err)
	}

	log.Printf("Loaded %d addresses from %s", len(addresses), filename)
	return addresses
}

// TO USE A DIFFERENT CONFIG:
// 1. Modify the DefaultConfig() function in config.go to return a different configuration
// 2. Or change the Network field in main.go to use a different network from the Networks map
// 3. For targeted scanning, use StoryTargetedScanConfig() or StoryAddressListConfig("addresses.txt")
