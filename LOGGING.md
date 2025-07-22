# CPIMP Scanner - Logging Control

## Log Levels

The scanner supports three log levels for performance optimization:

- **ERROR** (0): Only critical errors and failures
- **INFO** (1): Essential progress information and summaries (default)
- **DEBUG** (2): Detailed debugging information including API responses

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

## Recommended Usage

- **Production scanning**: `LOG_LEVEL=ERROR` for maximum speed
- **Monitoring progress**: `LOG_LEVEL=INFO` (default)
- **Debugging issues**: `LOG_LEVEL=DEBUG` to see API responses and detailed flow

## Log Output Examples

### ERROR Level
```
[ERROR] SKIPPED 0x123...: API error - failed to get address info
[ERROR] Error fetching logs for blocks 1000-2000: rate limit exceeded
```

### INFO Level
```
[INFO] Starting CPIMP Scanner with log level: 1
[INFO] Processing 84 addresses for creation blocks...
[INFO] VALID PROXY CONTRACT 0x123...: created in block 12345
[INFO] SUMMARY: Found 42 valid proxy contracts out of 84 addresses processed
```

### DEBUG Level
```
[DEBUG] Processing address 5/84: 0x123...
[DEBUG] Address API Response for 0x123...: {"is_contract": true, ...}
[DEBUG] Parsed for 0x123...: is_contract=true, creation_tx=0xabc...
[DEBUG] SKIPPED 0x456...: Not a smart contract (is_contract: false)
``` 