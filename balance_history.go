package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// Logging levels
const (
	LOG_ERROR = 0
	LOG_INFO  = 1
	LOG_DEBUG = 2
)

var logLevel = LOG_INFO // Default to INFO

// Logging functions
func logError(format string, args ...interface{}) {
	if logLevel >= LOG_ERROR {
		fmt.Printf("[ERROR] "+format+"\n", args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LOG_INFO {
		fmt.Printf("[INFO] "+format+"\n", args...)
	}
}

func logDebug(format string, args ...interface{}) {
	if logLevel >= LOG_DEBUG {
		fmt.Printf("[DEBUG] "+format+"\n", args...)
	}
}

// TokenTransfer represents a token transfer from the API
type TokenTransfer struct {
	BlockHash       string    `json:"block_hash"`
	LogIndex        int       `json:"log_index"`
	Method          string    `json:"method"`
	Timestamp       time.Time `json:"timestamp"`
	TransactionHash string    `json:"transaction_hash"`
	Type            string    `json:"type"`
	From            struct {
		Hash string `json:"hash"`
		Name string `json:"name"`
	} `json:"from"`
	To struct {
		Hash string `json:"hash"`
		Name string `json:"name"`
	} `json:"to"`
	Token struct {
		Address  string `json:"address"`
		Decimals string `json:"decimals"`
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
		Type     string `json:"type"`
	} `json:"token"`
	Total struct {
		Decimals string `json:"decimals"`
		Value    string `json:"value"`
	} `json:"total"`
}

// TokenTransfersResponse represents the API response
type TokenTransfersResponse struct {
	Items          []TokenTransfer `json:"items"`
	NextPageParams struct {
		BlockNumber int `json:"block_number"`
		Index       int `json:"index"`
	} `json:"next_page_params"`
}

// CoinBalanceHistory represents native coin balance history
type CoinBalanceHistory struct {
	TransactionHash string    `json:"transaction_hash"`
	BlockNumber     int       `json:"block_number"`
	BlockTimestamp  time.Time `json:"block_timestamp"`
	Delta           string    `json:"delta"`
	Value           string    `json:"value"`
}

// CoinBalanceResponse represents the coin balance history API response
type CoinBalanceResponse struct {
	Items          []CoinBalanceHistory `json:"items"`
	NextPageParams struct {
		BlockNumber int `json:"block_number"`
		ItemsCount  int `json:"items_count"`
	} `json:"next_page_params"`
}

// HistoryEntry represents a single row in our output CSV
type HistoryEntry struct {
	Address         string
	TokenAddress    string
	TokenSymbol     string
	TokenName       string
	Timestamp       time.Time
	BlockNumber     int
	TransactionHash string
	TransferType    string // "in", "out", "native_change"
	Amount          string // Decimal amount
	Balance         string // Running balance (calculated)
	FromAddress     string
	ToAddress       string
}

// convertToDecimal converts a raw token amount to decimal format
func convertToDecimal(rawAmount string, decimals string) string {
	if decimals == "" || rawAmount == "" {
		return rawAmount
	}

	decimalsInt, err := strconv.ParseInt(decimals, 10, 64)
	if err != nil {
		return rawAmount
	}

	// Handle negative amounts (for deltas)
	isNegative := strings.HasPrefix(rawAmount, "-")
	if isNegative {
		rawAmount = rawAmount[1:]
	}

	bigInt := new(big.Int)
	bigInt.SetString(rawAmount, 10)

	// Shift the decimal point
	shifted := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimalsInt), nil)

	// Use big.Rat for decimal division to preserve fractional parts
	bigRat := new(big.Rat).SetFrac(bigInt, shifted)

	result := bigRat.FloatString(6) // 6 decimal places precision

	if isNegative {
		result = "-" + result
	}

	return result
}

// loadLinesFromFile loads lines from a text file, skipping comments and empty lines
func loadLinesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var result []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		// Extract just the content part (remove any comments)
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if idx := strings.Index(line, "//"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if line != "" {
			result = append(result, line)
		}
	}

	return result, nil
}

// fetchTokenTransfers fetches all token transfers for an address and specific token
func fetchTokenTransfers(blockscoutURL, address, tokenAddress string) ([]TokenTransfer, error) {
	var allTransfers []TokenTransfer
	url := fmt.Sprintf("%s/api/v2/addresses/%s/token-transfers", blockscoutURL, address)

	// Don't filter by token here - get all transfers and filter later
	// This approach works better with the StoryScan API

	for {
		logDebug("Fetching token transfers: %s", url)

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch token transfers: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		var response TokenTransfersResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}

		// Filter transfers for the specific token if requested
		if tokenAddress != "" {
			var filteredTransfers []TokenTransfer
			for _, transfer := range response.Items {
				if strings.EqualFold(transfer.Token.Address, tokenAddress) {
					filteredTransfers = append(filteredTransfers, transfer)
				}
			}
			allTransfers = append(allTransfers, filteredTransfers...)
			logDebug("Got %d transfers for token %s (total: %d)", len(filteredTransfers), tokenAddress, len(allTransfers))
		} else {
			allTransfers = append(allTransfers, response.Items...)
			logDebug("Got %d transfers (total: %d)", len(response.Items), len(allTransfers))
		}

		// Check if there's a next page
		if response.NextPageParams.BlockNumber == 0 && response.NextPageParams.Index == 0 {
			break // No more pages
		}

		// Construct next page URL
		baseURL := fmt.Sprintf("%s/api/v2/addresses/%s/token-transfers", blockscoutURL, address)
		url = fmt.Sprintf("%s?block_number=%d&index=%d",
			baseURL, response.NextPageParams.BlockNumber, response.NextPageParams.Index)

		// Rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	return allTransfers, nil
}

// fetchCoinBalanceHistory fetches all native coin balance history for an address
func fetchCoinBalanceHistory(blockscoutURL, address string) ([]CoinBalanceHistory, error) {
	var allHistory []CoinBalanceHistory
	url := fmt.Sprintf("%s/api/v2/addresses/%s/coin-balance-history", blockscoutURL, address)

	for {
		logDebug("Fetching coin balance history: %s", url)

		resp, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch coin balance history: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %v", err)
		}

		var response CoinBalanceResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %v", err)
		}

		allHistory = append(allHistory, response.Items...)
		logDebug("Got %d balance changes (total: %d)", len(response.Items), len(allHistory))

		// Check if there's a next page
		if response.NextPageParams.BlockNumber == 0 && response.NextPageParams.ItemsCount == 0 {
			break // No more pages
		}

		// Construct next page URL
		url = fmt.Sprintf("%s/api/v2/addresses/%s/coin-balance-history?block_number=%d&items_count=%d",
			blockscoutURL, address, response.NextPageParams.BlockNumber, response.NextPageParams.ItemsCount)

		// Rate limiting
		time.Sleep(100 * time.Millisecond)
	}

	return allHistory, nil
}

func main() {
	// Check for LOG_LEVEL environment variable
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		switch strings.ToUpper(envLogLevel) {
		case "ERROR":
			logLevel = LOG_ERROR
		case "INFO":
			logLevel = LOG_INFO
		case "DEBUG":
			logLevel = LOG_DEBUG
		}
	}

	if len(os.Args) < 3 {
		fmt.Println("Usage: go run balance_history.go <addresses.txt> <tokens.txt> [network]")
		fmt.Println("Example: go run balance_history.go addresses.txt tokens.txt story")
		fmt.Println("Available networks: story, base, ethereum, polygon, arbitrum")
		fmt.Println("Set LOG_LEVEL environment variable to ERROR, INFO, or DEBUG (default: INFO)")
		os.Exit(1)
	}

	addressFile := os.Args[1]
	tokenFile := os.Args[2]
	network := "story" // default
	if len(os.Args) >= 4 {
		network = os.Args[3]
	}

	// Define networks locally
	networks := map[string]string{
		"story":    "https://www.storyscan.io",
		"base":     "https://base.blockscout.com",
		"ethereum": "https://eth.blockscout.com",
		"polygon":  "https://polygon.blockscout.com",
		"arbitrum": "https://arbitrum.blockscout.com",
	}

	blockscoutURL, exists := networks[network]
	if !exists {
		logError("Unknown network '%s'", network)
		fmt.Println("Available networks: story, base, ethereum, polygon, arbitrum")
		os.Exit(1)
	}

	logInfo("Balance History Tracker for %s network", network)
	logInfo("Using Blockscout API: %s", blockscoutURL)
	logInfo("Addresses file: %s", addressFile)
	logInfo("Tokens file: %s", tokenFile)

	// Load addresses
	addresses, err := loadLinesFromFile(addressFile)
	if err != nil {
		logError("Error loading addresses: %v", err)
		os.Exit(1)
	}

	// Load tokens
	tokens, err := loadLinesFromFile(tokenFile)
	if err != nil {
		logError("Error loading tokens: %v", err)
		os.Exit(1)
	}

	logInfo("Found %d addresses and %d tokens to analyze", len(addresses), len(tokens))

	// Create output CSV file
	outputFile := fmt.Sprintf("balance_history_%s.csv", network)
	file, err := os.Create(outputFile)
	if err != nil {
		logError("Error creating CSV file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{
		"address", "token_address", "token_symbol", "token_name",
		"timestamp", "block_number", "transaction_hash", "transfer_type",
		"amount", "from_address", "to_address",
	})

	var allEntries []HistoryEntry

	// Process each address
	for i, address := range addresses {
		// Always show progress for remaining addresses
		fmt.Printf("Processing address %d/%d: %s (%d remaining)\n", i+1, len(addresses), address, len(addresses)-(i+1))

		// Fetch native coin balance history
		logDebug("Fetching native balance history...")
		coinHistory, err := fetchCoinBalanceHistory(blockscoutURL, address)
		if err != nil {
			logError("Error fetching coin history: %v", err)
		} else {
			for _, entry := range coinHistory {
				historyEntry := HistoryEntry{
					Address:         address,
					TokenAddress:    "native",
					TokenSymbol:     "ETH", // Adjust based on network
					TokenName:       "Native Token",
					Timestamp:       entry.BlockTimestamp,
					BlockNumber:     entry.BlockNumber,
					TransactionHash: entry.TransactionHash,
					TransferType:    "native_change",
					Amount:          convertToDecimal(entry.Delta, "18"),
					FromAddress:     "",
					ToAddress:       "",
				}
				allEntries = append(allEntries, historyEntry)
			}
		}

		// Process each token for this address
		for j, tokenAddress := range tokens {
			logDebug("Fetching token transfers %d/%d: %s", j+1, len(tokens), tokenAddress)

			transfers, err := fetchTokenTransfers(blockscoutURL, address, tokenAddress)
			if err != nil {
				logError("Error fetching transfers: %v", err)
				continue
			}

			for _, transfer := range transfers {
				transferType := "unknown"
				amount := transfer.Total.Value

				// Determine if this is incoming or outgoing
				if strings.EqualFold(transfer.To.Hash, address) {
					transferType = "in"
				} else if strings.EqualFold(transfer.From.Hash, address) {
					transferType = "out"
					// Make outgoing amounts negative
					if !strings.HasPrefix(amount, "-") {
						amount = "-" + amount
					}
				}

				historyEntry := HistoryEntry{
					Address:         address,
					TokenAddress:    transfer.Token.Address,
					TokenSymbol:     transfer.Token.Symbol,
					TokenName:       transfer.Token.Name,
					Timestamp:       transfer.Timestamp,
					BlockNumber:     0, // Not provided in token transfers response
					TransactionHash: transfer.TransactionHash,
					TransferType:    transferType,
					Amount:          convertToDecimal(amount, transfer.Token.Decimals),
					FromAddress:     transfer.From.Hash,
					ToAddress:       transfer.To.Hash,
				}
				allEntries = append(allEntries, historyEntry)
			}
		}

		logDebug("Completed address %s", address)
	}

	// Write all entries to CSV
	logInfo("Writing %d entries to CSV...", len(allEntries))
	for _, entry := range allEntries {
		writer.Write([]string{
			entry.Address,
			entry.TokenAddress,
			entry.TokenSymbol,
			entry.TokenName,
			entry.Timestamp.Format(time.RFC3339),
			strconv.Itoa(entry.BlockNumber),
			entry.TransactionHash,
			entry.TransferType,
			entry.Amount,
			entry.FromAddress,
			entry.ToAddress,
		})
		writer.Flush()
	}

	fmt.Printf("✅ Balance history analysis complete!\n")
	fmt.Printf("📁 Results saved to: %s\n", outputFile)
	fmt.Printf("📊 Total entries: %d\n", len(allEntries))
	fmt.Printf("\n💡 Import this CSV into a spreadsheet to create balance history graphs:\n")
	fmt.Printf("   - Sort by address + token_address + timestamp\n")
	fmt.Printf("   - Create a running balance column: =SUM(previous_balance + amount)\n")
	fmt.Printf("   - Graph timestamp vs running balance for each address/token pair\n")
}
