#!/bin/bash

echo "CPIMP Scanner - Blockchain Event Scanner"
echo "========================================"

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go from https://golang.org/dl/"
    exit 1
fi

# Initialize Go module if go.mod doesn't exist
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module..."
    go mod init cpimp-scanner
fi

# Run the scanner
echo "Starting blockchain scan..."
echo "This may take a long time depending on the blockchain size."
echo "Press Ctrl+C to stop the scan if needed."
echo ""

go run *.go

echo ""
echo "Scan completed! Check the generated CSV file for results." 