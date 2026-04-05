# Stockyard Fleet

**Self-hosted vehicle and fleet management**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted tools.

## Quick Start

```bash
curl -fsSL https://stockyard.dev/tools/fleet/install.sh | sh
```

Or with Docker:

```bash
docker run -p 9809:9809 -v fleet_data:/data ghcr.io/stockyard-dev/stockyard-fleet
```

Open `http://localhost:9809` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9809` | HTTP port |
| `DATA_DIR` | `./fleet-data` | SQLite database directory |
| `STOCKYARD_LICENSE_KEY` | *(empty)* | License key for unlimited use |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 5 records | Unlimited |
| Price | Free | Included in bundle or $29.99/mo individual |

Get a license at [stockyard.dev](https://stockyard.dev).

## License

Apache 2.0
