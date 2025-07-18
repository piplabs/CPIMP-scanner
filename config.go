package main

import "time"

// Configuration for different blockchain networks
type NetworkConfig struct {
	Name          string
	BlockscoutURL string
	ExplorerURL   string
}

var Networks = map[string]NetworkConfig{
	"base": {
		Name:          "Base",
		BlockscoutURL: "https://base.blockscout.com",
		ExplorerURL:   "https://base.blockscout.com",
	},
	"ethereum": {
		Name:          "Ethereum",
		BlockscoutURL: "https://eth.blockscout.com",
		ExplorerURL:   "https://eth.blockscout.com",
	},
	"polygon": {
		Name:          "Polygon",
		BlockscoutURL: "https://polygon.blockscout.com",
		ExplorerURL:   "https://polygon.blockscout.com",
	},
	"optimism": {
		Name:          "Optimism",
		BlockscoutURL: "https://optimism.blockscout.com",
		ExplorerURL:   "https://optimism.blockscout.com",
	},
	"story": {
		Name:          "Story",
		BlockscoutURL: "https://www.storyscan.io",
		ExplorerURL:   "https://www.storyscan.io",
	},
}

// Scanner configuration
type ScannerConfig struct {
	// Network to scan (use key from Networks map)
	Network string

	// Event signature hash for Upgraded(address)
	// Default: keccak256("Upgraded(address)")
	EventTopic string

	// Block range for each API call (to avoid timeouts)
	BlockRange uint64

	// Rate limiting delay between API calls
	RateLimit time.Duration

	// Starting block (0 for genesis)
	StartBlock uint64

	// Ending block (0 for latest)
	EndBlock uint64

	// Output CSV filename
	OutputFile string

	// Specific addresses to scan (empty = scan all addresses)
	// When provided, only events from these addresses will be checked
	TargetAddresses []string
}

// Default configuration - uses Story network
func DefaultConfig() ScannerConfig {
	// return StoryAddressListConfig("eco_projects.txt")
	return EthereumNetworkListConfig("eco_projects.txt")
}
