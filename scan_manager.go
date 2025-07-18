package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ListActiveScans shows all ongoing scans
func ListActiveScans() {
	files, err := filepath.Glob("scan_progress_*.json")
	if err != nil {
		log.Printf("Error listing scan files: %v", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No active scans found.")
		return
	}

	fmt.Printf("Found %d active scan(s):\n\n", len(files))

	for _, file := range files {
		progress := loadProgress(file)
		if progress.ScanID == "" {
			continue
		}

		// Calculate progress percentage
		totalBlocks := progress.EndBlock - progress.StartBlock + 1
		processedBlocks := progress.CurrentBlock - progress.StartBlock
		percentage := float64(processedBlocks) / float64(totalBlocks) * 100

		fmt.Printf("Scan ID: %s\n", progress.ScanID)
		fmt.Printf("  Network: %s\n", progress.Network)
		fmt.Printf("  Progress: %.1f%% (%d/%d blocks)\n", percentage, processedBlocks, totalBlocks)
		fmt.Printf("  Current Block: %d\n", progress.CurrentBlock)
		fmt.Printf("  Target Addresses: %d\n", progress.AddressCount)
		fmt.Printf("  Logs Found: %d\n", progress.TotalLogs)
		fmt.Printf("  Duplicate Transactions: %d\n", progress.DuplicateTxs)
		fmt.Printf("  Last Updated: %s\n", progress.LastUpdated.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Progress File: %s\n", file)
		fmt.Println()
	}
}

// CleanupOldScans removes progress files older than specified duration
func CleanupOldScans(olderThan time.Duration) {
	files, err := filepath.Glob("scan_progress_*.json")
	if err != nil {
		log.Printf("Error listing scan files: %v", err)
		return
	}

	if len(files) == 0 {
		fmt.Println("No scan progress files found.")
		return
	}

	cleaned := 0
	cutoff := time.Now().Add(-olderThan)

	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			continue
		}

		if fileInfo.ModTime().Before(cutoff) {
			err = os.Remove(file)
			if err != nil {
				log.Printf("Error removing old scan file %s: %v", file, err)
			} else {
				fmt.Printf("Removed old scan progress file: %s\n", file)
				cleaned++
			}
		}
	}

	if cleaned == 0 {
		fmt.Printf("No scan progress files older than %v found.\n", olderThan)
	} else {
		fmt.Printf("Cleaned up %d old scan progress file(s).\n", cleaned)
	}
}

// DeleteScan removes a specific scan by ID
func DeleteScan(scanID string) {
	progressFile := getProgressFileName(scanID)

	if _, err := os.Stat(progressFile); os.IsNotExist(err) {
		fmt.Printf("Scan ID %s not found.\n", scanID)
		return
	}

	err := os.Remove(progressFile)
	if err != nil {
		log.Printf("Error removing scan %s: %v", scanID, err)
		return
	}

	fmt.Printf("Removed scan %s (file: %s)\n", scanID, progressFile)
}

// ShowScanDetails displays detailed information about a specific scan
func ShowScanDetails(scanID string) {
	progressFile := getProgressFileName(scanID)
	progress := loadProgress(progressFile)

	if progress.ScanID == "" {
		fmt.Printf("Scan ID %s not found.\n", scanID)
		return
	}

	// Calculate progress metrics
	totalBlocks := progress.EndBlock - progress.StartBlock + 1
	processedBlocks := progress.CurrentBlock - progress.StartBlock
	percentage := float64(processedBlocks) / float64(totalBlocks) * 100

	fmt.Printf("=== Scan Details ===\n")
	fmt.Printf("Scan ID: %s\n", progress.ScanID)
	fmt.Printf("Network: %s\n", progress.Network)
	fmt.Printf("Event Topic: %s\n", progress.EventTopic)
	fmt.Printf("Target Addresses: %d\n", progress.AddressCount)
	fmt.Printf("Block Range: %d - %d (%d total)\n", progress.StartBlock, progress.EndBlock, totalBlocks)
	fmt.Printf("Current Block: %d\n", progress.CurrentBlock)
	fmt.Printf("Progress: %.2f%% (%d/%d blocks)\n", percentage, processedBlocks, totalBlocks)
	fmt.Printf("Logs Found: %d\n", progress.TotalLogs)
	fmt.Printf("Duplicate Transactions: %d\n", progress.DuplicateTxs)
	fmt.Printf("Processed Transactions: %d\n", progress.ProcessedTxs)
	fmt.Printf("Last Updated: %s\n", progress.LastUpdated.Format("2006-01-02 15:04:05 MST"))
	fmt.Printf("Progress File: %s\n", progressFile)

	// Estimate completion time if scan is active
	if processedBlocks > 0 && processedBlocks < totalBlocks {
		elapsed := time.Since(progress.LastUpdated)
		if elapsed < 24*time.Hour { // Only estimate if recently updated
			remainingBlocks := totalBlocks - processedBlocks
			timePerBlock := elapsed / time.Duration(processedBlocks)
			estimatedTimeRemaining := timePerBlock * time.Duration(remainingBlocks)
			fmt.Printf("Estimated Time Remaining: %v\n", estimatedTimeRemaining.Truncate(time.Second))
		}
	}
}

// Helper function to get scan ID from partial ID
func findScanByPartialID(partialID string) string {
	files, err := filepath.Glob("scan_progress_*.json")
	if err != nil {
		return ""
	}

	for _, file := range files {
		// Extract scan ID from filename
		filename := filepath.Base(file)
		if strings.HasPrefix(filename, "scan_progress_") && strings.HasSuffix(filename, ".json") {
			scanID := filename[14 : len(filename)-5] // Remove "scan_progress_" and ".json"
			if strings.HasPrefix(scanID, partialID) {
				return scanID
			}
		}
	}
	return ""
}
