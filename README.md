# eth-balance-exporter

A Prometheus exporter for monitoring Ethereum wallet balances across multiple RPC endpoints.

## Description

This exporter connects to one or more Ethereum RPC endpoints and monitors the balance of specified wallet addresses. It exposes the balance metrics in Prometheus format at the `/metrics` endpoint, making it easy to integrate with Prometheus, Grafana, and other monitoring tools.

## Features

- Monitor multiple wallet addresses across different RPC endpoints
- Automatic Wei to ETH conversion
- Client connection caching for better performance
- Support for both HTTP and HTTPS RPC endpoints
- Prometheus-compatible metrics format
- Lightweight Docker image (~14MB content size)

## Requirements

- Go 1.25.6 or later (for building from source)
- Docker (for running containerized version)
- Access to Ethereum RPC endpoint(s) (e.g., Infura, Alchemy, local node)

## Installation

### From Source

```bash
go build -o eth-balance-exporter
```

### Docker

Build the Docker image:
```bash
docker build -t eth-balance-exporter:latest .
```

Or pull from registry (if published):
```bash
docker pull your-registry/eth-balance-exporter:latest
```

## Configuration

The exporter is configured via environment variables:

### Environment Variables

| Variable | Required | Description | Format |
|----------|----------|-------------|--------|
| `RPC_URL_MAPPING` | Yes | Maps RPC URLs to wallet addresses | `RPC_URL:wallet1,wallet2\|RPC_URL2:wallet3` |

### RPC_URL_MAPPING Format

The format allows you to specify multiple RPC endpoints with their associated wallet addresses:

```
RPC_URL_1:wallet_address_1,wallet_address_2|RPC_URL_2:wallet_address_3,wallet_address_4
```

- Multiple RPC URLs are separated by pipe (`|`)
- Each RPC URL is followed by a colon (`:`)
- Multiple wallet addresses for the same RPC are separated by comma (`,`)
- RPC URLs must start with `http://` or `https://`

## Usage

### Running the Binary

```bash
export RPC_URL_MAPPING="https://mainnet.infura.io/v3/YOUR_API_KEY:0x742d35Cc6634C0532925a3b844Bc454e4438f44e,0x123..."
./eth-balance-exporter
```

### Running with Docker

```bash
docker run -d \
  -p 8080:8080 \
  -e RPC_URL_MAPPING="https://mainnet.infura.io/v3/YOUR_API_KEY:0x742d35Cc6634C0532925a3b844Bc454e4438f44e" \
  --name eth-balance-exporter \
  eth-balance-exporter:latest
```

### Docker Compose Example

```yaml
version: '3.8'
services:
  eth-balance-exporter:
    image: eth-balance-exporter:latest
    ports:
      - "8080:8080"
    environment:
      - RPC_URL_MAPPING=https://mainnet.infura.io/v3/YOUR_API_KEY:0x742d35Cc6634C0532925a3b844Bc454e4438f44e
    restart: unless-stopped
```

## Examples

### Single RPC with One Wallet

```bash
export RPC_URL_MAPPING="https://mainnet.infura.io/v3/YOUR_API_KEY:0x742d35Cc6634C0532925a3b844Bc454e4438f44e"
./eth-balance-exporter
```

### Single RPC with Multiple Wallets

```bash
export RPC_URL_MAPPING="https://mainnet.infura.io/v3/YOUR_API_KEY:0x742d35Cc6634C0532925a3b844Bc454e4438f44e,0x123...,0x456..."
./eth-balance-exporter
```

### Multiple RPCs with Multiple Wallets

```bash
export RPC_URL_MAPPING="https://mainnet.infura.io/v3/YOUR_API_KEY:0x742d35Cc6634C0532925a3b844Bc454e4438f44e|https://polygon-rpc.com:0x123...,0x456..."
./eth-balance-exporter
```

## Metrics

### Exposed Metrics

The exporter exposes the following metrics at `http://localhost:8080/metrics`:

```
# HELP wallet_balance_eth Balance of the specified wallet in ETH
# TYPE wallet_balance_eth gauge
wallet_balance_eth{wallet="0x742d35Cc6634C0532925a3b844Bc454e4438f44e"} 1.234567
```

### Metric Details

- **Name**: `wallet_balance_eth`
- **Type**: Gauge
- **Labels**:
  - `wallet`: The Ethereum wallet address
- **Value**: Balance in ETH (converted from Wei, where 1 ETH = 10^18 Wei)

## Prometheus Configuration

Add this job to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'eth-balance-exporter'
    static_configs:
      - targets: ['localhost:8080']
```

## Health Check

To verify the exporter is running:

```bash
curl http://localhost:8080/metrics
```

## Logging

The exporter logs the following events:
- Successful RPC connections
- Failed balance retrievals
- Connection errors to RPC endpoints

Logs are written to stdout/stderr and can be viewed with:

```bash
# For Docker
docker logs eth-balance-exporter

# For binary
./eth-balance-exporter 2>&1 | tee exporter.log
```

## Troubleshooting

### Error: "RPC_URL_MAPPING environment variable must be set"

Make sure you've set the `RPC_URL_MAPPING` environment variable before running the exporter.

### Error: "invalid format in RPC_URL_MAPPING"

Check that your RPC_URL_MAPPING follows the correct format:
- RPC URLs must start with `http://` or `https://`
- Use `:` to separate RPC URL from wallet addresses
- Use `,` to separate multiple wallet addresses
- Use `|` to separate multiple RPC URL configurations

### Connection Errors

If you see RPC connection errors:
- Verify your RPC endpoint is accessible
- Check your API key is valid (for hosted services)
- Ensure network connectivity to the RPC endpoint
- Check firewall rules if using a local node

## License

[Add your license here]

## Contributing

[Add contribution guidelines here]
