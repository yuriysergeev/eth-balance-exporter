package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// WalletBalanceCollector collects and exposes balance metrics for multiple wallets across RPC URLs.
type WalletBalanceCollector struct {
	rpcWalletMapping map[string][]string
	clientCache      map[string]*ethclient.Client
	balanceMetric    *prometheus.Desc
	mutex            sync.Mutex
}

// NewWalletBalanceCollector creates a new WalletBalanceCollector.
func NewWalletBalanceCollector(rpcWalletMapping map[string][]string) *WalletBalanceCollector {
	return &WalletBalanceCollector{
		rpcWalletMapping: rpcWalletMapping,
		clientCache:      make(map[string]*ethclient.Client),
		balanceMetric: prometheus.NewDesc(
			"wallet_balance_eth",
			"Balance of the specified wallet in ETH",
			[]string{"wallet"},
			nil,
		),
	}
}

// Describe sends the descriptor of the metric to Prometheus.
func (c *WalletBalanceCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.balanceMetric
}

// Collect fetches the balance for each wallet and sends it to Prometheus.
func (c *WalletBalanceCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for rpcURL, wallets := range c.rpcWalletMapping {
		client, err := c.getClient(rpcURL)
		if err != nil {
			log.Printf("Error connecting to RPC URL %s: %v", rpcURL, err)
			continue
		}

		for _, walletAddress := range wallets {
			balance, err := c.getWalletBalance(client, walletAddress)
			if err != nil {
				log.Printf("Error retrieving balance for wallet %s: %v", walletAddress, err)
				continue
			}

			ch <- prometheus.MustNewConstMetric(
				c.balanceMetric,
				prometheus.GaugeValue,
				balance,
				walletAddress,
			)
		}
	}
}

// getClient retrieves or creates an ethclient.Client for the given RPC URL.
func (c *WalletBalanceCollector) getClient(rpcURL string) (*ethclient.Client, error) {
	if client, exists := c.clientCache[rpcURL]; exists {
		return client, nil
	}

	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	c.clientCache[rpcURL] = client
	log.Printf("Successfully connected to RPC URL: %s", rpcURL)
	return client, nil
}

// getWalletBalance retrieves the balance of the wallet.
func (c *WalletBalanceCollector) getWalletBalance(client *ethclient.Client, walletAddress string) (float64, error) {
	address := common.HexToAddress(walletAddress)
	balanceWei, err := client.BalanceAt(context.Background(), address, nil)
	if err != nil {
		return 0, err
	}

	// Convert Wei to ETH (1 ETH = 10^18 Wei)
	balanceETH := new(big.Float).Quo(new(big.Float).SetInt(balanceWei), big.NewFloat(1e18))
	balance, _ := balanceETH.Float64()
	return balance, nil
}

// parseRPCMapping parses the RPC_URL_MAPPING environment variable into a map of RPC URLs and associated wallet addresses.
func parseRPCMapping(rpcMapping string) (map[string][]string, error) {
	rpcWalletMapping := make(map[string][]string)
	mappings := strings.Split(rpcMapping, "|")

	for _, mapping := range mappings {
		mapping = strings.TrimSpace(mapping)

		// Locate the first colon after 'http://' or 'https://'
		colonIndex := strings.Index(mapping, ":")
		if strings.HasPrefix(mapping, "http://") {
			colonIndex = strings.Index(mapping[7:], ":") + 7
		} else if strings.HasPrefix(mapping, "https://") {
			colonIndex = strings.Index(mapping[8:], ":") + 8
		}

		if colonIndex == -1 || colonIndex == len(mapping)-1 {
			return nil, fmt.Errorf("invalid format in RPC_URL_MAPPING: %s (missing colon or wallets)", mapping)
		}

		rpcURL := mapping[:colonIndex]
		wallets := mapping[colonIndex+1:]

		// Validate RPC URL format
		if !strings.HasPrefix(rpcURL, "http://") && !strings.HasPrefix(rpcURL, "https://") {
			return nil, fmt.Errorf("invalid RPC URL: %s (must start with http:// or https://)", rpcURL)
		}

		// Split wallet addresses into a slice
		walletList := strings.Split(wallets, ",")
		rpcWalletMapping[rpcURL] = walletList
	}

	return rpcWalletMapping, nil
}

func main() {
	// Get RPC_URL_MAPPING from environment variables
	rpcMapping := os.Getenv("RPC_URL_MAPPING")
	if rpcMapping == "" {
		log.Fatal("RPC_URL_MAPPING environment variable must be set")
	}

	rpcWalletMapping, err := parseRPCMapping(rpcMapping)
	if err != nil {
		log.Fatalf("Error parsing RPC_URL_MAPPING: %v", err)
	}

	// Create and register the Prometheus collector
	collector := NewWalletBalanceCollector(rpcWalletMapping)
	prometheus.MustRegister(collector)

	// Expose metrics at /metrics
	http.Handle("/metrics", promhttp.Handler())

	// Start the HTTP server
	port := "8080"
	log.Printf("Starting server on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
