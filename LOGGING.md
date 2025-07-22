# CPIMP Scanner - Logging Control

## Log Levels

The scanner supports three log levels for performance optimization:

- **ERROR** (0): Only critical errors, progress indicators, and completion summary
- **INFO** (1): Essential progress information and scan details (default)
- **DEBUG** (2): Detailed debugging information including API responses and transaction logs

## Setting Log Level

### Environment Variable
```bash
export LOG_LEVEL=ERROR    # Minimal output for maximum speed
export LOG_LEVEL=INFO     # Normal operation (default)
export LOG_LEVEL=DEBUG    # Detailed debugging

./run_scanner.sh
```

### Inline with Runner Script
```bash
LOG_LEVEL=ERROR ./run_scanner.sh    # Fastest performance
LOG_LEVEL=INFO ./run_scanner.sh     # Normal operation
LOG_LEVEL=DEBUG ./run_scanner.sh    # Full debugging
```

## Performance Impact

- **ERROR**: Fastest execution, minimal I/O overhead
- **INFO**: Good balance of performance and monitoring
- **DEBUG**: Slowest due to extensive logging, only use for troubleshooting

## Progress Indicators

Progress indicators work at **all log levels** (including ERROR) and show:

- Address processing progress (every 10 addresses)
- Current scanning status with counts and percentages
- Time estimates and completion predictions
- Valid vs skipped contract counts

These use `fmt.Printf` instead of the logging system, so they're always visible for monitoring.

### Progress Output Examples
```
📊 Progress: 20/84 (23.8%) | Valid: 15 | Skipped: 5
🚀 Starting scan: 15 addresses total, 0 already completed
📍 Scanning address 3/15 (12 remaining): 0x1234...abcd
✅ Address 0x1234...abcd complete: 42 logs, 3 duplicate transactions
📊 Overall Progress: 3/15 (20.0%) | Remaining: 12 addresses
⏱️  Estimated time remaining: 2h15m30s (avg: 11m17s per address)
```

## What Shows at Each Level

### ERROR Level (Maximum Speed)
```
📊 Progress: 20/84 (23.8%) | Valid: 15 | Skipped: 5
🚀 Starting scan: 15 addresses total, 0 already completed
📍 Scanning address 3/15 (12 remaining): 0x1234...abcd
✅ Address 0x1234...abcd complete: 42 logs, 3 duplicate transactions
📊 Overall Progress: 3/15 (20.0%) | Remaining: 12 addresses
⏱️  Estimated time remaining: 2h15m30s (avg: 11m17s per address)
=== Scan Complete (ID: abc123) ===
Total time: 2h34m15s
Results saved to: ethereum_address_list_scan.csv
[ERROR] Failed to get address info: API returned status 500
```

### INFO Level (Normal Operation)
```
[INFO] Starting CPIMP Scanner with log level: 1
[INFO] Processing 84 addresses for creation blocks...
[INFO] VALID PROXY CONTRACT 0x123...: created in block 12345
[INFO] SUMMARY: Found 42 valid proxy contracts out of 84 addresses processed
[INFO] Starting from creation block: 12345
[INFO] Scanning blocks 12345 to 15000 for 0x123...
[INFO] Previous progress: 150 logs found, 12 duplicate transactions
+ All ERROR level output
+ Detailed scan results and API timing
```

### DEBUG Level (Full Debugging)
```
[DEBUG] Processing address 5/84: 0x123...
[DEBUG] Address API Response for 0x123...: {"is_contract": true, ...}
[DEBUG] Block transaction details:
[DEBUG]     Block 12345: 3 transactions
[DEBUG]       TX 1: 0xabc123...
[DEBUG]         2 logs/events:
[DEBUG]           Log 1: address=0x123...
+ All INFO level output
+ Complete transaction and event details
```

## Recommended Usage

- **Production scanning**: `LOG_LEVEL=ERROR` for maximum speed
- **Monitoring progress**: `LOG_LEVEL=INFO` (default)
- **Debugging issues**: `LOG_LEVEL=DEBUG` to see API responses and detailed flow 