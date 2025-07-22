#!/bin/bash

# Runner script for the CPIMP scanner
SCANNER_DIR="$(pwd)"
LOG_FILE="$SCANNER_DIR/scanner.log"
SCREEN_NAME="cpimp-scanner"

cd $SCANNER_DIR

echo "🔄 Starting CPIMP Scanner..."
echo "📁 Working directory: $(pwd)"
echo "📝 Logs will be written to: $LOG_FILE"

# Kill existing screen session if it exists
screen -S $SCREEN_NAME -X quit 2>/dev/null

# Set log level (INFO for normal operation, DEBUG for troubleshooting, ERROR for minimal output)
export LOG_LEVEL=${LOG_LEVEL:-INFO}

# Start scanner in screen session with logging
screen -S $SCREEN_NAME -dm bash -c "
    echo '🚀 Scanner started at: $(date) with LOG_LEVEL=$LOG_LEVEL' | tee -a $LOG_FILE
    LOG_LEVEL=$LOG_LEVEL go run main.go config.go network_configs.go 2>&1 | tee -a $LOG_FILE
"

echo "✅ Scanner started in screen session: $SCREEN_NAME"
echo "📊 Log level: $LOG_LEVEL"
echo ""
echo "📋 Useful commands:"
echo "  View logs:      tail -f $LOG_FILE"
echo "  Attach screen:  screen -r $SCREEN_NAME"
echo "  List screens:   screen -ls"
echo "  Kill scanner:   screen -S $SCREEN_NAME -X quit"
echo ""
echo "🔍 Monitor progress with:"
echo "  watch -n 30 'tail -20 $LOG_FILE'"
echo ""
echo "🎛️  Log level control:"
echo "  Minimal output: LOG_LEVEL=ERROR ./run_scanner.sh"
echo "  Normal output:  LOG_LEVEL=INFO ./run_scanner.sh"
echo "  Debug output:   LOG_LEVEL=DEBUG ./run_scanner.sh" 