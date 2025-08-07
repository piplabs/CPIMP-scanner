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

// ProxyInfo represents the result for each address
type ProxyInfo struct {
	ProxyAddress          string
	ImplementationAddress string
	ImplementationName    string
	ContractName          string
	PublicTags            string
	NativeBalance         string
	USDCBalance           string
	WIPBalance            string
	WETHBalance           string
	VIPBalance            string
	StIPBalance           string
}

// AddressInfoResponse represents the Blockscout API response for address info
type AddressInfoResponse struct {
	IsContract      bool   `json:"is_contract"`
	Name            string `json:"name"`
	CoinBalance     string `json:"coin_balance"`
	Implementations []struct {
		Address string `json:"address"`
		Name    string `json:"name"`
	} `json:"implementations"`
	PublicTags []struct {
		DisplayName string `json:"display_name"`
		Label       string `json:"label"`
	} `json:"public_tags"`
	Token struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
		Type   string `json:"type"`
	} `json:"token"`
}

// TokenBalance represents a token balance from the API
type TokenBalance struct {
	Token struct {
		Address  string `json:"address"`
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Decimals string `json:"decimals"`
	} `json:"token"`
	Value string `json:"value"`
}

// TokenBalancesResponse represents the response from token-balances endpoint
type TokenBalancesResponse []TokenBalance

// MetadataResponse represents the Blockscout metadata service response
type MetadataResponse struct {
	Addresses map[string]struct {
		Tags []struct {
			Slug    string `json:"slug"`
			Name    string `json:"name"`
			TagType string `json:"tagType"`
			Ordinal int    `json:"ordinal"`
			Meta    string `json:"meta"`
		} `json:"tags"`
	} `json:"addresses"`
}

// getAddressInfo fetches address information from Blockscout API
func getAddressInfo(blockscoutURL, address string) (*AddressInfoResponse, error) {
	url := fmt.Sprintf("%s/api/v2/addresses/%s", blockscoutURL, address)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch address info: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var addressInfo AddressInfoResponse
	if err := json.Unmarshal(body, &addressInfo); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &addressInfo, nil
}

// getMetadataTags fetches additional tags from Blockscout metadata service
func getMetadataTags(address string) ([]string, error) {
	// Story network chainId is 1514
	url := fmt.Sprintf("https://metadata.services.blockscout.com/api/v1/metadata?addresses=%s&chainId=1514&tagsLimit=20", address)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch metadata: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("metadata API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata response: %v", err)
	}

	var metadata MetadataResponse
	if err := json.Unmarshal(body, &metadata); err != nil {
		return nil, fmt.Errorf("failed to parse metadata response: %v", err)
	}

	// Extract tag names
	var tagNames []string
	if addressData, exists := metadata.Addresses[address]; exists {
		for _, tag := range addressData.Tags {
			tagNames = append(tagNames, tag.Name)
		}
	}

	return tagNames, nil
}

// getTokenBalances fetches ERC20 token balances for an address
func getTokenBalances(blockscoutURL, address string) (map[string]string, error) {
	url := fmt.Sprintf("%s/api/v2/addresses/%s/token-balances", blockscoutURL, address)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch token balances: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("token balances API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token balances response: %v", err)
	}

	var balances TokenBalancesResponse
	if err := json.Unmarshal(body, &balances); err != nil {
		return nil, fmt.Errorf("failed to parse token balances response: %v", err)
	}

	// Create a map of token addresses to balances
	tokenBalances := make(map[string]string)
	for _, balance := range balances {
		tokenBalances[strings.ToLower(balance.Token.Address)] = balance.Value
	}

	return tokenBalances, nil
}

// convertRawBalanceToDecimal converts a raw token balance (string) to a decimal string
func convertRawBalanceToDecimal(rawBalance string, decimals string) string {
	if decimals == "" {
		return rawBalance // Return raw if decimals are not available
	}

	decimalsInt, err := strconv.ParseInt(decimals, 10, 64)
	if err != nil {
		return rawBalance // Return raw if decimals are invalid
	}

	bigInt := new(big.Int)
	bigInt.SetString(rawBalance, 10)

	// Shift the decimal point by the number of decimals
	shifted := new(big.Int).Exp(big.NewInt(10), big.NewInt(decimalsInt), nil)
	decimalValue := new(big.Int).Div(bigInt, shifted)

	return decimalValue.String()
}

// checkProxy checks if an address is a proxy and returns implementation address
func checkProxy(blockscoutURL, address string) ProxyInfo {
	result := ProxyInfo{
		ProxyAddress:          address,
		ImplementationAddress: "not a proxy",
		ImplementationName:    "",
		ContractName:          "",
		PublicTags:            "",
		NativeBalance:         "0",
		USDCBalance:           "0",
		WIPBalance:            "0",
		WETHBalance:           "0",
		VIPBalance:            "0",
		StIPBalance:           "0",
	}

	// Define Story network token addresses
	tokenAddresses := map[string]string{
		"usdc": "0xf1815bd50389c46847f0bda824ec8da914045d14",
		"wip":  "0x1514000000000000000000000000000000000000",
		"weth": "0xbab93b7ad7fe8692a878b95a8e689423437cc500",
		"vip":  "0x5267f7ee069ceb3d8f1c760c215569b79d0685ad",
		"stip": "0xd07faed671decf3c5a6cc038dad97c8efdb507c0",
	}

	// Get address info from Blockscout API
	addressInfo, err := getAddressInfo(blockscoutURL, address)
	if err != nil {
		fmt.Printf("❌ Error checking %s: %v\n", address, err)
		result.ImplementationAddress = "error"
		return result
	}

	// Extract native balance
	result.NativeBalance = convertRawBalanceToDecimal(addressInfo.CoinBalance, "18") // Native token has 18 decimals

	// Extract contract name (primary source)
	result.ContractName = addressInfo.Name

	// If no contract name but has token info, use token name
	if result.ContractName == "" && addressInfo.Token.Name != "" {
		result.ContractName = fmt.Sprintf("%s (%s)", addressInfo.Token.Name, addressInfo.Token.Symbol)
	}

	// Extract public tags from main API
	var tagNames []string
	for _, tag := range addressInfo.PublicTags {
		if tag.DisplayName != "" {
			tagNames = append(tagNames, tag.DisplayName)
		}
	}

	// Get additional tags from metadata service
	metadataTags, err := getMetadataTags(address)
	if err != nil {
		fmt.Printf("⚠️ Could not fetch metadata tags for %s: %v\n", address, err)
	} else {
		tagNames = append(tagNames, metadataTags...)
	}

	result.PublicTags = strings.Join(tagNames, "; ")

	// Get token balances
	tokenBalances, err := getTokenBalances(blockscoutURL, address)
	if err != nil {
		fmt.Printf("⚠️ Could not fetch token balances for %s: %v\n", address, err)
	} else {
		// Map specific token balances
		if balance, exists := tokenBalances[strings.ToLower(tokenAddresses["usdc"])]; exists {
			result.USDCBalance = convertRawBalanceToDecimal(balance, "6") // USDC has 6 decimals
		}
		if balance, exists := tokenBalances[strings.ToLower(tokenAddresses["wip"])]; exists {
			result.WIPBalance = convertRawBalanceToDecimal(balance, "18") // WIP has 18 decimals
		}
		if balance, exists := tokenBalances[strings.ToLower(tokenAddresses["weth"])]; exists {
			result.WETHBalance = convertRawBalanceToDecimal(balance, "18") // WETH has 18 decimals
		}
		if balance, exists := tokenBalances[strings.ToLower(tokenAddresses["vip"])]; exists {
			result.VIPBalance = convertRawBalanceToDecimal(balance, "18") // VIP has 18 decimals
		}
		if balance, exists := tokenBalances[strings.ToLower(tokenAddresses["stip"])]; exists {
			result.StIPBalance = convertRawBalanceToDecimal(balance, "18") // StIP has 18 decimals
		}
	}

	// Check if it's a contract
	if !addressInfo.IsContract {
		fmt.Printf("⏭️ %s: Not a contract\n", address)
		result.ImplementationAddress = "not a contract"
		return result
	}

	// Check if it has implementations (proxy pattern)
	if len(addressInfo.Implementations) == 0 {
		fmt.Printf("⏭️ %s: Contract but not a proxy\n", address)
		return result
	}

	// Get the first implementation address and name
	implementation := addressInfo.Implementations[0].Address
	implementationName := addressInfo.Implementations[0].Name

	result.ImplementationAddress = implementation
	result.ImplementationName = implementationName

	displayName := implementationName
	if displayName == "" {
		displayName = "unnamed"
	}

	fmt.Printf("✅ %s: Proxy -> %s (%s)\n", address, implementation, displayName)
	return result
}

// loadAddressesFromFile loads addresses from a text file (copy from network_configs.go)
func loadAddressesFromFile(filename string) []string {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file %s: %v\n", filename, err)
		return nil
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", filename, err)
		return nil
	}

	lines := strings.Split(string(content), "\n")
	var addresses []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}
		// Extract just the address part (remove any comments)
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if idx := strings.Index(line, "//"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if line != "" {
			addresses = append(addresses, line)
		}
	}

	return addresses
}

// getAvailableNetworks returns a formatted string of available networks
func getAvailableNetworks() string {
	// Define networks locally to avoid import conflicts
	networks := map[string]string{
		"story":    "https://www.storyscan.io",
		"base":     "https://base.blockscout.com",
		"ethereum": "https://eth.blockscout.com",
		"polygon":  "https://polygon.blockscout.com",
		"arbitrum": "https://arbitrum.blockscout.com",
	}

	var networkNames []string
	for network := range networks {
		networkNames = append(networkNames, network)
	}
	return strings.Join(networkNames, ", ")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run proxy_checker.go <address_file.txt> [network]")
		fmt.Println("Example: go run proxy_checker.go eco_projects.txt story")
		fmt.Println("Available networks:", getAvailableNetworks())
		os.Exit(1)
	}

	addressFile := os.Args[1]
	network := "story" // default
	if len(os.Args) >= 3 {
		network = os.Args[2]
	}

	// Define networks locally to avoid import conflicts with main scanner
	networks := map[string]string{
		"story":    "https://www.storyscan.io",
		"base":     "https://base.blockscout.com",
		"ethereum": "https://eth.blockscout.com",
		"polygon":  "https://polygon.blockscout.com",
		"arbitrum": "https://arbitrum.blockscout.com",
	}

	// Get network configuration
	blockscoutURL, exists := networks[network]
	if !exists {
		fmt.Printf("Error: Unknown network '%s'\n", network)
		fmt.Println("Available networks:", getAvailableNetworks())
		os.Exit(1)
	}

	fmt.Printf("🔍 Proxy Checker for %s network\n", network)
	fmt.Printf("📡 Using Blockscout API: %s\n", blockscoutURL)
	fmt.Printf("📁 Loading addresses from: %s\n", addressFile)

	// Load addresses from file
	addresses := loadAddressesFromFile(addressFile)
	if addresses == nil {
		fmt.Printf("Error loading addresses from %s\n", addressFile)
		os.Exit(1)
	}

	fmt.Printf("📋 Found %d addresses to check\n\n", len(addresses))

	// Create output CSV file
	outputFile := fmt.Sprintf("proxy_analysis_%s.csv", network)
	file, err := os.Create(outputFile)
	if err != nil {
		fmt.Printf("Error creating CSV file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Write CSV header
	writer.Write([]string{"proxy_address", "implementation_address", "implementation_name", "contract_name", "public_tags", "native_balance", "usdc_balance", "wip_balance", "weth_balance", "vip_balance", "stip_balance"})

	// Process each address
	results := make([]ProxyInfo, 0, len(addresses))
	for i, address := range addresses {
		fmt.Printf("📍 Checking %d/%d: %s\n", i+1, len(addresses), address)

		result := checkProxy(blockscoutURL, address)
		results = append(results, result)

		// Write to CSV immediately
		writer.Write([]string{result.ProxyAddress, result.ImplementationAddress, result.ImplementationName, result.ContractName, result.PublicTags, result.NativeBalance, result.USDCBalance, result.WIPBalance, result.WETHBalance, result.VIPBalance, result.StIPBalance})
		writer.Flush() // Ensure data is written

		// Rate limiting to avoid overwhelming the API
		if i < len(addresses)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Print summary
	fmt.Printf("\n📊 Summary:\n")
	proxies := 0
	notProxies := 0
	notContracts := 0
	errors := 0

	for _, result := range results {
		switch result.ImplementationAddress {
		case "not a proxy":
			notProxies++
		case "not a contract":
			notContracts++
		case "error":
			errors++
		default:
			proxies++
		}
	}

	fmt.Printf("✅ Proxies found: %d\n", proxies)
	fmt.Printf("⏭️ Not proxies: %d\n", notProxies)
	fmt.Printf("⏭️ Not contracts: %d\n", notContracts)
	fmt.Printf("❌ Errors: %d\n", errors)
	fmt.Printf("📁 Results saved to: %s\n", outputFile)
}
