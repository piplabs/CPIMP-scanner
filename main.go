package main

import (
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
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

var logLevel = LOG_INFO // Default to INFO level

func logDebug(format string, args ...interface{}) {
	if logLevel >= LOG_DEBUG {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func logInfo(format string, args ...interface{}) {
	if logLevel >= LOG_INFO {
		log.Printf("[INFO] "+format, args...)
	}
}

func logError(format string, args ...interface{}) {
	if logLevel >= LOG_ERROR {
		log.Printf("[ERROR] "+format, args...)
	}
}

type LogEntry struct {
	TransactionHash string `json:"transactionHash"`
	BlockNumber     string `json:"blockNumber"`
	Address         string `json:"address"`
}

type Transaction struct {
	From string `json:"from"`
	Hash string `json:"hash"`
}

type ApiResponse struct {
	Result []LogEntry `json:"result"`
}

type TransactionResponse struct {
	Result Transaction `json:"result"`
}

type BlockResponse struct {
	Result struct {
		Number string `json:"number"`
	} `json:"result"`
}

type ProgressTracker struct {
	ScanID       string    `json:"scan_id"`
	Network      string    `json:"network"`
	EventTopic   string    `json:"event_topic"`
	AddressCount int       `json:"address_count"`
	StartBlock   uint64    `json:"start_block"`
	CurrentBlock uint64    `json:"current_block"`
	EndBlock     uint64    `json:"end_block"`
	LastUpdated  time.Time `json:"last_updated"`
	TotalLogs    int       `json:"total_logs"`
	DuplicateTxs int       `json:"duplicate_txs"`
	ProcessedTxs int       `json:"processed_txs"`
}

// Generate unique scan ID based on configuration
func generateScanID(config ScannerConfig) string {
	hasher := sha256.New()

	// Hash network and event topic
	hasher.Write([]byte(config.Network))
	hasher.Write([]byte(config.EventTopic))

	// Hash target addresses (sort first for consistency)
	if len(config.TargetAddresses) > 0 {
		sortedAddresses := make([]string, len(config.TargetAddresses))
		copy(sortedAddresses, config.TargetAddresses)
		sort.Strings(sortedAddresses)

		for _, addr := range sortedAddresses {
			hasher.Write([]byte(strings.ToLower(addr)))
		}
	} else {
		// Use a special marker for "all addresses" scans
		hasher.Write([]byte("__ALL_ADDRESSES__"))
	}

	// Hash block range to distinguish different scans
	hasher.Write([]byte(fmt.Sprintf("%d-%d", config.StartBlock, config.EndBlock)))

	hash := hasher.Sum(nil)
	return hex.EncodeToString(hash)[:16] // Use first 16 chars for readability
}

// Get progress file name for this specific scan
func getProgressFileName(scanID string) string {
	return fmt.Sprintf("scan_progress_%s.json", scanID)
}

// ContractInfo holds information about a contract
type ContractInfo struct {
	Address       string `json:"address"`
	CreationBlock uint64 `json:"creation_block"`
	CreationTx    string `json:"creation_tx"`
	Processed     bool   `json:"processed"`
}

// AddressProgress tracks progress for individual addresses
type AddressProgress struct {
	Addresses    map[string]ContractInfo `json:"addresses"`
	ScanID       string                  `json:"scan_id"`
	Network      string                  `json:"network"`
	EventTopic   string                  `json:"event_topic"`
	LastUpdated  time.Time               `json:"last_updated"`
	TotalLogs    int                     `json:"total_logs"`
	DuplicateTxs int                     `json:"duplicate_txs"`
	ProcessedTxs int                     `json:"processed_txs"`
}

// getContractCreationBlock fetches the creation block for a contract address using Blockscout v2 API
func getContractCreationBlock(blockscoutURL, address string) (uint64, string, error) {
	// First, get the contract info to find creation transaction hash
	url := fmt.Sprintf("%s/api/v2/addresses/%s", blockscoutURL, address)

	resp, err := http.Get(url)
	if err != nil {
		return 0, "", fmt.Errorf("failed to fetch contract info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, "", fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, "", fmt.Errorf("failed to read response: %v", err)
	}

	// Parse address response
	var addressInfo struct {
		CreationTransactionHash string `json:"creation_transaction_hash"`
		IsContract              bool   `json:"is_contract"`
		ProxyType               string `json:"proxy_type"`
		Implementations         []struct {
			Address string `json:"address"`
		} `json:"implementations"`
	}

	// Debug: Log the raw API response
	logDebug("Address API Response for %s: %s", address, string(body))

	if err := json.Unmarshal(body, &addressInfo); err != nil {
		return 0, "", fmt.Errorf("failed to parse address response: %v", err)
	}

	// Debug: Log parsed fields
	logDebug("Parsed for %s: is_contract=%t, creation_tx=%s, implementations=%d",
		address, addressInfo.IsContract, addressInfo.CreationTransactionHash, len(addressInfo.Implementations))

	// Check if this is a smart contract
	if !addressInfo.IsContract {
		return 0, "", fmt.Errorf("address is not a smart contract (is_contract: false)")
	}

	// Check if this is a proxy contract (has implementations array with at least one entry)
	if len(addressInfo.Implementations) == 0 {
		return 0, "", fmt.Errorf("not a proxy contract (no implementations found)")
	}

	// Check for valid creation transaction hash
	if addressInfo.CreationTransactionHash == "" {
		return 0, "", fmt.Errorf("no creation transaction found")
	}

	// Now get the transaction details to find the block number
	blockNumber, err := getTransactionBlockNumber(blockscoutURL, addressInfo.CreationTransactionHash)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get transaction block: %v", err)
	}

	return blockNumber, addressInfo.CreationTransactionHash, nil
}

// getTransactionBlockNumber gets the block number for a transaction hash using Blockscout v2 API
func getTransactionBlockNumber(blockscoutURL, txHash string) (uint64, error) {
	url := fmt.Sprintf("%s/api/v2/transactions/%s", blockscoutURL, txHash)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to fetch transaction info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse transaction response
	var txInfo struct {
		BlockNumber int64 `json:"block_number"`
	}

	if err := json.Unmarshal(body, &txInfo); err != nil {
		return 0, fmt.Errorf("failed to parse transaction response: %v", err)
	}

	if txInfo.BlockNumber <= 0 {
		return 0, fmt.Errorf("invalid block number: %d", txInfo.BlockNumber)
	}

	return uint64(txInfo.BlockNumber), nil
}

// processAddressCreationBlocks processes each address individually to get creation blocks
func processAddressCreationBlocks(blockscoutURL string, targetAddresses []string) map[string]ContractInfo {
	addressInfo := make(map[string]ContractInfo)

	logInfo("Processing %d addresses for creation blocks...", len(targetAddresses))

	// Progress tracking
	totalAddresses := len(targetAddresses)
	validContracts := 0
	skippedContracts := 0

	for i, address := range targetAddresses {
		// Show progress every 10 addresses or at key milestones (always shown regardless of log level)
		if i%10 == 0 || i == totalAddresses-1 {
			progress := float64(i+1) / float64(totalAddresses) * 100
			fmt.Printf("📊 Progress: %d/%d (%.1f%%) | Valid: %d | Skipped: %d\n",
				i+1, totalAddresses, progress, validContracts, skippedContracts)
		}
		logDebug("Processing address %d/%d: %s", i+1, len(targetAddresses), address)

		creationBlock, creationTx, err := getContractCreationBlock(blockscoutURL, address)
		if err != nil {
			skippedContracts++

			// Always log skipped addresses (minimal info)
			if strings.Contains(err.Error(), "is_contract: false") {
				fmt.Printf("⏭️  SKIP %s: Not a contract\n", address)
				logDebug("SKIPPED %s: Not a smart contract (is_contract: false)", address)
			} else if strings.Contains(err.Error(), "no implementations found") {
				fmt.Printf("⏭️  SKIP %s: Not a proxy\n", address)
				logDebug("SKIPPED %s: Not a proxy contract (no implementations found)", address)
			} else if strings.Contains(err.Error(), "no creation transaction") {
				fmt.Printf("⏭️  SKIP %s: No creation tx\n", address)
				logDebug("SKIPPED %s: No creation transaction found", address)
			} else if strings.Contains(err.Error(), "API returned status") {
				fmt.Printf("❌ SKIP %s: API error\n", address)
				logError("SKIPPED %s: API error - %v", address, err)
			} else {
				fmt.Printf("❌ SKIP %s: Error\n", address)
				logError("SKIPPED %s: Failed to get creation info - %v", address, err)
			}
			// Don't include filtered addresses in the addressInfo map
			continue
		} else {
			validContracts++

			// Always log valid addresses (minimal info)
			fmt.Printf("✅ VALID %s: Block %d\n", address, creationBlock)
			logInfo("VALID PROXY CONTRACT %s: created in block %d (tx: %s)", address, creationBlock, creationTx)
			addressInfo[address] = ContractInfo{
				Address:       address,
				CreationBlock: creationBlock,
				CreationTx:    creationTx,
				Processed:     false,
			}
		}

		// Add delay to avoid rate limiting
		time.Sleep(200 * time.Millisecond)
	}

	logInfo("SUMMARY: Found %d valid proxy contracts out of %d addresses processed", len(addressInfo), len(targetAddresses))
	return addressInfo
}

// loadAddressProgress loads progress for address-based scanning
func loadAddressProgress(progressFile string) AddressProgress {
	file, err := os.Open(progressFile)
	if err != nil {
		return AddressProgress{
			Addresses: make(map[string]ContractInfo),
		}
	}
	defer file.Close()

	var progress AddressProgress
	if err := json.NewDecoder(file).Decode(&progress); err != nil {
		logError("Warning: Could not decode progress file: %v", err)
		return AddressProgress{
			Addresses: make(map[string]ContractInfo),
		}
	}

	return progress
}

// saveAddressProgress saves progress for address-based scanning
func saveAddressProgress(progressFile string, progress AddressProgress) {
	progress.LastUpdated = time.Now()

	file, err := os.Create(progressFile)
	if err != nil {
		logError("Warning: Could not save progress: %v", err)
		return
	}
	defer file.Close()

	if err := json.NewEncoder(file).Encode(progress); err != nil {
		logError("Warning: Could not encode progress: %v", err)
	}
}

// getBlockTransactions fetches all transactions in a block
func getBlockTransactions(blockscoutURL string, blockNumber uint64) ([]string, error) {
	url := fmt.Sprintf("%s/api/v2/blocks/%d/transactions", blockscoutURL, blockNumber)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch block transactions: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse transactions response
	var txResponse struct {
		Items []struct {
			Hash string `json:"hash"`
		} `json:"items"`
	}

	if err := json.Unmarshal(body, &txResponse); err != nil {
		return nil, fmt.Errorf("failed to parse transactions response: %v", err)
	}

	var txHashes []string
	for _, tx := range txResponse.Items {
		txHashes = append(txHashes, tx.Hash)
	}

	return txHashes, nil
}

// getTransactionLogs fetches all logs/events for a specific transaction
func getTransactionLogs(blockscoutURL string, txHash string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v2/transactions/%s/logs", blockscoutURL, txHash)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch transaction logs: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse logs response
	var logsResponse struct {
		Items []map[string]interface{} `json:"items"`
	}

	if err := json.Unmarshal(body, &logsResponse); err != nil {
		return nil, fmt.Errorf("failed to parse logs response: %v", err)
	}

	return logsResponse.Items, nil
}

func main() {
	// Set log level from environment variable
	if level := os.Getenv("LOG_LEVEL"); level != "" {
		switch strings.ToUpper(level) {
		case "ERROR":
			logLevel = LOG_ERROR
		case "INFO":
			logLevel = LOG_INFO
		case "DEBUG":
			logLevel = LOG_DEBUG
		default:
			logLevel = LOG_INFO
		}
	}

	logInfo("Starting CPIMP Scanner with log level: %d", logLevel)

	// Load configuration
	config := DefaultConfig()

	// Generate unique scan ID
	scanID := generateScanID(config)
	progressFile := getProgressFileName(scanID)

	// Get network configuration
	network, exists := Networks[config.Network]
	if !exists {
		log.Fatalf("Unknown network: %s", config.Network)
	}

	fmt.Printf("Starting blockchain scan for Upgraded events on %s...\n", network.Name)
	fmt.Printf("Scan ID: %s\n", scanID)

	// Get the latest block number
	latestBlock, err := getLatestBlockNumber(network.BlockscoutURL)
	if err != nil {
		log.Fatalf("Failed to get latest block number: %v", err)
	}

	// Load address-based progress
	addressProgress := loadAddressProgress(progressFile)

	// Initialize or update address progress
	if addressProgress.ScanID == "" {
		// Fresh scan - process addresses to get creation blocks
		addressInfo := processAddressCreationBlocks(network.BlockscoutURL, config.TargetAddresses)

		addressProgress = AddressProgress{
			Addresses:   addressInfo,
			ScanID:      scanID,
			Network:     config.Network,
			EventTopic:  config.EventTopic,
			LastUpdated: time.Now(),
		}

		saveAddressProgress(progressFile, addressProgress)

		// Determine the earliest creation block for overall scan range
		var earliestBlock uint64 = ^uint64(0) // Max uint64
		validContracts := 0

		for _, info := range addressInfo {
			if info.CreationBlock > 0 {
				validContracts++
				if info.CreationBlock < earliestBlock {
					earliestBlock = info.CreationBlock
				}
			}
		}

		startBlock := config.StartBlock
		if validContracts > 0 && (config.StartBlock == 0 || earliestBlock < config.StartBlock) {
			startBlock = earliestBlock
		}

		endBlock := config.EndBlock
		if endBlock == 0 {
			endBlock = latestBlock
		}

		fmt.Printf("Starting fresh address-based scan from block %d to %d (latest: %d)\n", startBlock, endBlock, latestBlock)
		fmt.Printf("Targeting %d addresses (%d with known creation blocks)\n", len(config.TargetAddresses), validContracts)
	} else {
		fmt.Printf("Resuming address-based scan\n")
		fmt.Printf("📋 Address Summary: %d total loaded, %d valid proxy contracts found\n",
			len(config.TargetAddresses), len(addressProgress.Addresses))
		fmt.Printf("   (Only proxy contracts with implementations are scanned for Upgraded events)\n")

		if logLevel >= LOG_INFO {
			fmt.Printf("Previous progress: %d logs found, %d duplicate transactions\n", addressProgress.TotalLogs, addressProgress.DuplicateTxs)
		}

		// Show address status
		processed := 0
		for _, info := range addressProgress.Addresses {
			if info.Processed {
				processed++
			}
		}
		fmt.Printf("Address progress: %d/%d addresses completed\n", processed, len(addressProgress.Addresses))
	}

	fmt.Printf("Progress file: %s\n\n", progressFile)

	// Prepare CSV file
	file, err := os.OpenFile(config.OutputFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open CSV file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header only if file is empty
	fileInfo, _ := file.Stat()
	if fileInfo.Size() == 0 {
		writer.Write([]string{"Transaction Hash", "Explorer Link", "From Address", "Block Number"})
	}

	// Track performance metrics
	startTime := time.Now()
	var totalAPITime time.Duration
	requestCount := 0

	// Progress tracking for address scanning
	totalAddressesToScan := len(addressProgress.Addresses)
	completedAddresses := 0
	for _, info := range addressProgress.Addresses {
		if info.Processed {
			completedAddresses++
		}
	}

	fmt.Printf("\n🚀 Starting scan: %d addresses total, %d already completed\n", totalAddressesToScan, completedAddresses)

	// Scan each address individually from its creation block
	addressIndex := 0
	for address, info := range addressProgress.Addresses {
		addressIndex++

		if info.Processed {
			fmt.Printf("⏭️  Skipping already processed address %d/%d: %s\n", addressIndex, totalAddressesToScan, address)
			continue
		}

		// Show progress (always visible regardless of log level)
		remainingAddresses := totalAddressesToScan - completedAddresses
		fmt.Printf("\n📍 Scanning address %d/%d (%d remaining): %s\n",
			addressIndex, totalAddressesToScan, remainingAddresses, address)
		if info.CreationBlock > 0 {
			logInfo("Starting from creation block: %d", info.CreationBlock)
		} else {
			logInfo("Creation block unknown, starting from block 0")
		}

		startBlock := info.CreationBlock
		if startBlock == 0 {
			startBlock = config.StartBlock
		}

		endBlock := config.EndBlock
		if endBlock == 0 {
			endBlock = latestBlock
		}

		// Scan this address in chunks
		addressLogs := 0
		addressDuplicates := 0

		for fromBlock := startBlock; fromBlock <= endBlock; fromBlock += config.BlockRange {
			toBlock := fromBlock + config.BlockRange - 1
			if toBlock > endBlock {
				toBlock = endBlock
			}

			logDebug("Scanning blocks %d to %d for %s...", fromBlock, toBlock, address)

			// Log all transactions in each block in this range (DEBUG level only)
			if logLevel >= LOG_DEBUG {
				fmt.Printf("\n  Block transaction details:\n")
				for blockNum := fromBlock; blockNum <= toBlock; blockNum++ {
					txHashes, err := getBlockTransactions(network.BlockscoutURL, blockNum)
					if err != nil {
						fmt.Printf("    Block %d: Error fetching transactions - %v\n", blockNum, err)
						continue
					}

					if len(txHashes) == 0 {
						fmt.Printf("    Block %d: No transactions\n", blockNum)
					} else {
						fmt.Printf("    Block %d: %d transactions\n", blockNum, len(txHashes))
						for i, txHash := range txHashes {
							fmt.Printf("      TX %d: %s\n", i+1, txHash)

							// Fetch and display logs for this transaction
							logs, err := getTransactionLogs(network.BlockscoutURL, txHash)
							if err != nil {
								fmt.Printf("        Error fetching logs: %v\n", err)
								continue
							}

							if len(logs) == 0 {
								fmt.Printf("        No logs/events\n")
							} else {
								fmt.Printf("        %d logs/events:\n", len(logs))
								for j, logEntry := range logs {
									// Extract key fields from the log
									address := ""
									topics := []interface{}{}
									data := ""
									decoded := map[string]interface{}{}

									if addr, ok := logEntry["address"].(map[string]interface{}); ok {
										if hash, exists := addr["hash"].(string); exists {
											address = hash
										}
									}
									if topicsArray, ok := logEntry["topics"].([]interface{}); ok {
										topics = topicsArray
									}
									if dataStr, ok := logEntry["data"].(string); ok {
										data = dataStr
									}
									if decodedData, ok := logEntry["decoded"].(map[string]interface{}); ok {
										decoded = decodedData
									}

									fmt.Printf("          Log %d: address=%s\n", j+1, address)

									// Show decoded event information if available
									if methodCall, exists := decoded["method_call"].(string); exists {
										fmt.Printf("                 event=%s\n", methodCall)
										if params, exists := decoded["parameters"].([]interface{}); exists {
											fmt.Printf("                 parameters:\n")
											for k, param := range params {
												if paramMap, ok := param.(map[string]interface{}); ok {
													name := paramMap["name"]
													value := paramMap["value"]
													paramType := paramMap["type"]
													indexed := paramMap["indexed"]
													fmt.Printf("                   %d. %s (%s, indexed:%v) = %v\n", k+1, name, paramType, indexed, value)
												}
											}
										}
									} else {
										// Fallback to raw topic display
										fmt.Printf("                 topics=%v\n", topics)
									}

									if len(data) > 100 {
										fmt.Printf("                 data=%s... (%d chars)\n", data[:100], len(data))
									} else if data != "" && data != "0x" {
										fmt.Printf("                 data=%s\n", data)
									}
								}
							}
						}
					}
				}
			} // End of DEBUG level transaction logging

			// Measure API call time
			apiStart := time.Now()
			logs, err := fetchLogs(network.BlockscoutURL, config.EventTopic, fromBlock, toBlock, []string{address})
			apiDuration := time.Since(apiStart)
			totalAPITime += apiDuration
			requestCount++

			if err != nil {
				logError("Error fetching logs for blocks %d-%d: %v", fromBlock, toBlock, err)
				continue
			}

			// Group logs by transaction hash for this chunk
			chunkLogs := make(map[string][]LogEntry)
			for _, logEntry := range logs {
				chunkLogs[logEntry.TransactionHash] = append(chunkLogs[logEntry.TransactionHash], logEntry)
			}

			addressProgress.TotalLogs += len(logs)
			addressLogs += len(logs)

			// Log details about found events (DEBUG level only)
			if len(logs) > 0 && logLevel >= LOG_DEBUG {
				fmt.Printf("\n  Found %d Upgraded events in this chunk:\n", len(logs))
				for i, logEntry := range logs {
					fmt.Printf("    Event %d: tx=%s, block=%s, address=%s\n",
						i+1, logEntry.TransactionHash, logEntry.BlockNumber, logEntry.Address)
				}
			}

			// Process transactions with 2+ events in this chunk
			chunkDuplicates := 0
			for txHash, txLogs := range chunkLogs {
				if len(txLogs) >= 2 {
					chunkDuplicates++
					addressProgress.DuplicateTxs++
					addressDuplicates++

					// Only show duplicate details in DEBUG mode
					if logLevel >= LOG_DEBUG {
						fmt.Printf("\n  *** DUPLICATE FOUND *** Transaction %s has %d Upgraded events:\n", txHash, len(txLogs))
						for i, txLog := range txLogs {
							fmt.Printf("    Event %d: block=%s, address=%s\n", i+1, txLog.BlockNumber, txLog.Address)
						}
					}

					// Get transaction details
					fromAddress, err := getTransactionFrom(network.BlockscoutURL, txHash)
					if err != nil {
						logError("Error getting transaction details for %s: %v", txHash, err)
						fromAddress = "Unknown"
					}

					// Use block number from first log
					blockNumber := txLogs[0].BlockNumber

					// Construct explorer link
					explorerLink := fmt.Sprintf("%s/tx/%s", network.ExplorerURL, txHash)

					// Write to CSV
					writer.Write([]string{txHash, explorerLink, fromAddress, blockNumber})
					addressProgress.ProcessedTxs++

					// Rate limiting for transaction details
					time.Sleep(config.RateLimit)
				}
			}

			// Log chunk results (DEBUG level only)
			if logLevel >= LOG_DEBUG {
				avgAPITime := totalAPITime / time.Duration(requestCount)
				fmt.Printf(" Found %d logs, %d duplicate txs (avg API: %v)\n",
					len(logs), chunkDuplicates, avgAPITime.Truncate(time.Millisecond))
			}

			// Adaptive rate limiting
			if apiDuration > 500*time.Millisecond {
				time.Sleep(config.RateLimit * 2)
			} else {
				time.Sleep(config.RateLimit)
			}
		}

		// Mark address as processed and save progress
		info.Processed = true
		addressProgress.Addresses[address] = info
		saveAddressProgress(progressFile, addressProgress)

		completedAddresses++
		remainingAddresses = totalAddressesToScan - completedAddresses
		overallProgress := float64(completedAddresses) / float64(totalAddressesToScan) * 100

		fmt.Printf("✅ Address %s complete: %d logs, %d duplicate transactions\n", address, addressLogs, addressDuplicates)
		fmt.Printf("📊 Overall Progress: %d/%d (%.1f%%) | Remaining: %d addresses\n",
			completedAddresses, totalAddressesToScan, overallProgress, remainingAddresses)

		// Estimate time remaining
		if completedAddresses > 0 && remainingAddresses > 0 {
			elapsedSoFar := time.Since(startTime)
			avgTimePerAddress := elapsedSoFar / time.Duration(completedAddresses)
			estimatedTimeRemaining := avgTimePerAddress * time.Duration(remainingAddresses)
			fmt.Printf("⏱️  Estimated time remaining: %v (avg: %v per address)\n",
				estimatedTimeRemaining.Truncate(time.Second), avgTimePerAddress.Truncate(time.Second))
		}
	}

	// Final summary
	elapsed := time.Since(startTime)
	fmt.Printf("\n=== Scan Complete (ID: %s) ===\n", scanID)
	fmt.Printf("Total time: %v\n", elapsed.Truncate(time.Second))

	// Detailed results (INFO level and above)
	if logLevel >= LOG_INFO {
		fmt.Printf("Total logs found: %d\n", addressProgress.TotalLogs)
		fmt.Printf("Total transactions with 2+ Upgraded events: %d\n", addressProgress.DuplicateTxs)
		fmt.Printf("Total API calls: %d\n", requestCount)
		if requestCount > 0 {
			fmt.Printf("Average API response time: %v\n", (totalAPITime / time.Duration(requestCount)).Truncate(time.Millisecond))
		}
	}
	fmt.Printf("Results saved to: %s\n", config.OutputFile)

	// Clean up progress file on successful completion
	os.Remove(progressFile)
	fmt.Printf("Progress file %s removed (scan completed)\n", progressFile)
}

func getLatestBlockNumber(blockscoutURL string) (uint64, error) {
	// Try JSON-RPC format first (for Story network)
	url := fmt.Sprintf("%s/api?module=block&action=eth_block_number", blockscoutURL)

	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Try JSON-RPC format first
	var jsonRpcResp struct {
		JsonRpc string `json:"jsonrpc"`
		Result  string `json:"result"`
		Id      int    `json:"id"`
	}
	err = json.Unmarshal(body, &jsonRpcResp)
	if err == nil && jsonRpcResp.Result != "" {
		// Convert hex string to uint64 (remove 0x prefix)
		blockNumber, err := strconv.ParseUint(jsonRpcResp.Result[2:], 16, 64)
		if err != nil {
			return 0, err
		}
		return blockNumber, nil
	}

	// Fallback to Blockscout format
	var blockResp BlockResponse
	err = json.Unmarshal(body, &blockResp)
	if err != nil {
		return 0, err
	}

	// Convert hex string to uint64
	if len(blockResp.Result.Number) < 2 {
		return 0, fmt.Errorf("invalid block number format: %s", blockResp.Result.Number)
	}
	blockNumber, err := strconv.ParseUint(blockResp.Result.Number[2:], 16, 64)
	if err != nil {
		return 0, err
	}

	return blockNumber, nil
}

func fetchLogs(blockscoutURL, eventTopic string, fromBlock, toBlock uint64, targetAddresses []string) ([]LogEntry, error) {
	url := fmt.Sprintf("%s/api?module=logs&action=getLogs&fromBlock=%d&toBlock=%d&topic0=%s",
		blockscoutURL, fromBlock, toBlock, eventTopic)

	// Add address filter if target addresses are specified
	if len(targetAddresses) > 0 {
		// Blockscout API supports multiple addresses separated by commas
		addressList := strings.Join(targetAddresses, ",")
		url += "&address=" + addressList
	}

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		return nil, err
	}

	return apiResponse.Result, nil
}

func getTransactionFrom(blockscoutURL, txHash string) (string, error) {
	url := fmt.Sprintf("%s/api?module=proxy&action=eth_getTransactionByHash&txhash=%s", blockscoutURL, txHash)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var txResponse TransactionResponse
	err = json.Unmarshal(body, &txResponse)
	if err != nil {
		return "", err
	}

	return txResponse.Result.From, nil
}
