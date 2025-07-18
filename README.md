# CPIMP Scanner

A Go script that scans a blockchain using the Blockscout API to find transactions where the `Upgraded(address)` event was emitted multiple times (2 or more) within the same transaction.

## Prerequisites

- Go 1.21 or higher installed
- Internet connection to access the Blockscout API

## Configuration

The scanner is pre-configured to scan the **Story network** by default using https://www.storyscan.io/. 

To scan a different network, you can:

1. **Change the default network**: Modify the `DefaultConfig()` function in `config.go` to return a different network configuration (e.g., `BaseNetworkConfig()`, `EthereumNetworkConfig()`)

2. **Use available network configurations** in `network_configs.go`:
   - `StoryNetworkConfig()` - Story Protocol (default, scans all addresses)
   - `StoryTargetedScanConfig()` - Story Protocol with specific addresses (much faster)
   - `StoryAddressListConfig("addresses.txt")` - Story Protocol with addresses from file
   - `BaseNetworkConfig()` - Base network
   - `EthereumNetworkConfig()` - Ethereum mainnet
   - `RecentBlocksConfig()` - Scan only recent blocks
   - `FastScanConfig()` - Faster scanning with larger block ranges

3. **Supported networks** in the Networks map:
   - **story**: Story Protocol (https://www.storyscan.io/)
   - **base**: Base network (https://base.blockscout.com)
   - **ethereum**: Ethereum mainnet (https://eth.blockscout.com)
   - **polygon**: Polygon (https://polygon.blockscout.com)
   - **optimism**: Optimism (https://optimism.blockscout.com)

4. **Configuration options**:
   - **EventTopic**: The keccak256 hash of the `Upgraded(address)` event
   - **BlockRange**: Number of blocks to scan in each API call (default: 10,000)
   - **RateLimit**: Delay between API calls to respect rate limits (default: 500ms)
   - **TargetAddresses**: List of specific contract addresses to monitor (empty = all addresses)

## How to Run

1. **Initialize the Go module** (if not already done):
   ```bash
   go mod init cpimp-scanner
   ```

2. **Run the scanner**:
   ```bash
   go run main.go
   ```

## What the Script Does

1. **Fetches the latest block number** from the blockchain
2. **Scans the blockchain** in chunks (default: 10,000 blocks per chunk)
3. **Queries Blockscout API** for logs matching the `Upgraded(address)` event topic
   - If `TargetAddresses` is specified, only scans those specific contract addresses (much faster)
   - If `TargetAddresses` is empty, scans all addresses on the network
4. **Groups logs by transaction hash** and counts occurrences
5. **Identifies transactions** with 2 or more `Upgraded(address)` events
6. **Retrieves transaction details** to get the 'from' address
7. **Outputs results to CSV** with the following columns:
   - Transaction Hash
   - Explorer Link
   - From Address
   - Block Number

## Address Targeting Feature

For faster scanning, you can target specific contract addresses:

### Method 1: Hardcoded addresses
```go
// In config.go, change DefaultConfig() to:
func DefaultConfig() ScannerConfig {
    return StoryTargetedScanConfig()
}
```

### Method 2: Load from file
```go
// In config.go, change DefaultConfig() to:
func DefaultConfig() ScannerConfig {
    return StoryAddressListConfig("my_addresses.txt")
}
```

### Address file format
Create a text file (e.g., `my_addresses.txt`) with one address per line:
```
# Comments start with # or //
0x1234567890123456789012345678901234567890
0x2345678901234567890123456789012345678901
# Add as many addresses as needed
```

**Benefits of address targeting:**
- **10-100x faster** scanning (depending on how many addresses you target vs. total contracts)
- **Lower API usage** and reduced rate limiting
- **Focus on specific contracts** you care about

## Output

The script creates a file called `upgraded_transactions.csv` containing all transactions where the `Upgraded(address)` event was emitted multiple times.

## Performance Considerations

- **Full chain scan**: This script scans the entire blockchain from genesis block to latest
- **Rate limiting**: Built-in delays to respect API rate limits
- **Chunking**: Processes blocks in manageable chunks to avoid timeouts
- **Resume capability**: Automatic resume from interruption with unique progress tracking
- **Multiple scans**: Run different scans independently (different addresses, networks, etc.)

## Scan Management

Each scan gets a unique ID based on its configuration (network, addresses, event topic). This allows multiple independent scans:

### Progress Files
- **Unique IDs**: Each scan configuration gets a unique hash-based ID
- **Independent progress**: Different scans don't interfere with each other
- **Automatic cleanup**: Progress files are removed when scans complete

### Example Scan IDs
- `a1b2c3d4e5f6g7h8` - Story network, all addresses
- `f9e8d7c6b5a49382` - Story network, 3 specific addresses  
- `9876543210abcdef` - Base network, all addresses

### Multiple Simultaneous Scans
You can run different scans at the same time:
```bash
# Terminal 1: Scan all addresses
go run *.go

# Terminal 2: Scan specific addresses (different config)
# Modify config.go to use StoryTargetedScanConfig(), then:
go run *.go

# Each will have its own progress file and can resume independently
```

## Example Output

The CSV will contain entries like:
```
Transaction Hash,Explorer Link,From Address,Block Number
0x1234...5678,https://base.blockscout.com/tx/0x1234...5678,0xabcd...ef01,12345678
```

## Troubleshooting

- **API timeouts**: Reduce the `BLOCK_RANGE` constant
- **Rate limit errors**: Increase the `RATE_LIMIT_DELAY` constant
- **Network issues**: Check your internet connection and Blockscout URL
- **No results**: Verify the event topic hash is correct for your use case

## Customization

To scan for different events:
1. Calculate the keccak256 hash of your event signature
2. Update the `UPGRADED_EVENT_TOPIC` constant
3. Modify the struct definitions if your event has different parameters 