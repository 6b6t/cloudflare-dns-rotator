# cloudflare-dns-rotator

[![CI](https://github.com/6b6t/cloudflare-dns-rotator/actions/workflows/ci.yml/badge.svg)](https://github.com/6b6t/cloudflare-dns-rotator/actions/workflows/ci.yml)
[![Docker](https://github.com/6b6t/cloudflare-dns-rotator/actions/workflows/docker.yml/badge.svg)](https://github.com/6b6t/cloudflare-dns-rotator/actions/workflows/docker.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/6b6t/cloudflare-dns-rotator)](https://goreportcard.com/report/github.com/6b6t/cloudflare-dns-rotator)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Automatically rotate Cloudflare DNS CNAME records between multiple targets at configurable intervals. Perfect for load distribution, failover scenarios, or A/B testing with DNS-level routing.

## Features

- **Random rotation** - Automatically switches CNAME targets at configurable intervals
- **Smart selection** - Always picks a different target from the current one
- **Auto-creation** - Creates the DNS record if it doesn't exist
- **Graceful shutdown** - Handles SIGINT/SIGTERM for clean termination
- **Resilient** - Continues running even if individual rotations fail
- **Simple config** - JSON-based configuration file

## Installation

### Docker (recommended)

```bash
docker pull ghcr.io/6b6t/cloudflare-dns-rotator:latest
```

### From source

```bash
go install github.com/6b6t/cloudflare-dns-rotator@latest
```

### Build manually

```bash
git clone https://github.com/6b6t/cloudflare-dns-rotator.git
cd cloudflare-dns-rotator
go build -o cloudflare-dns-rotator .
```

## Configuration

Create a `config.json` file (see `config.example.json`):

```json
{
  "api_token": "your-cloudflare-api-token-here",
  "domain": "example.com",
  "record_name": "app",
  "targets": [
    "server1.backend.example.com",
    "server2.backend.example.com",
    "server3.backend.example.com"
  ],
  "interval": "5m",
  "proxied": false,
  "ttl": 1
}
```

### Configuration Options

| Option | Description | Required |
|--------|-------------|----------|
| `api_token` | Cloudflare API token with DNS edit permissions | Yes |
| `domain` | The domain/zone to manage (e.g., `example.com`) | Yes |
| `record_name` | CNAME record name without domain (e.g., `app` for `app.example.com`) | Yes |
| `targets` | List of target addresses to rotate between (minimum 2) | Yes |
| `interval` | Rotation interval (e.g., `30s`, `5m`, `1h`, `24h`) | Yes |
| `proxied` | Whether to proxy through Cloudflare (orange cloud) | No (default: `false`) |
| `ttl` | DNS TTL in seconds (1 = automatic) | No (default: `1`) |

### Cloudflare API Token

Create an API token at [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens) with the following permissions:

- **Zone > DNS > Edit** - For the zone you want to manage

## Usage

```bash
# Run with default config (./config.json)
./cloudflare-dns-rotator

# Run with custom config path
./cloudflare-dns-rotator -config /path/to/config.json
```

### Example Output

```
2025/01/15 10:00:00 main.go:20: Cloudflare DNS Rotator starting...
2025/01/15 10:00:00 main.go:21: Loading configuration from config.json
2025/01/15 10:00:00 main.go:27: Configuration loaded successfully
2025/01/15 10:00:00 main.go:28:   Domain: example.com
2025/01/15 10:00:00 main.go:29:   Record: app.example.com
2025/01/15 10:00:00 main.go:30:   Targets: [server1.backend.example.com server2.backend.example.com server3.backend.example.com]
2025/01/15 10:00:00 main.go:31:   Interval: 5m
2025/01/15 10:00:01 rotator.go:52: Found existing record app.example.com -> server1.backend.example.com
2025/01/15 10:00:01 main.go:53: Starting rotation loop...
2025/01/15 10:05:01 rotator.go:60: Rotating app.example.com: server1.backend.example.com -> server2.backend.example.com
2025/01/15 10:05:01 rotator.go:73: Successfully rotated to server2.backend.example.com
```

### Running as a Service

#### systemd

Create `/etc/systemd/system/cloudflare-dns-rotator.service`:

```ini
[Unit]
Description=Cloudflare DNS Rotator
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/cloudflare-dns-rotator -config /etc/cloudflare-dns-rotator/config.json
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Then enable and start:

```bash
sudo systemctl enable cloudflare-dns-rotator
sudo systemctl start cloudflare-dns-rotator
```

#### Docker

Run the official image from GitHub Container Registry:

```bash
docker run -d \
  --name cloudflare-dns-rotator \
  --restart unless-stopped \
  -v $(pwd)/config.json:/config/config.json:ro \
  ghcr.io/6b6t/cloudflare-dns-rotator:latest
```

#### Docker Compose

```yaml
services:
  cloudflare-dns-rotator:
    image: ghcr.io/6b6t/cloudflare-dns-rotator:latest
    container_name: cloudflare-dns-rotator
    restart: unless-stopped
    volumes:
      - ./config.json:/config/config.json:ro
```

```bash
docker compose up -d
```

## Use Cases

- **Load distribution** - Spread traffic across multiple backends at the DNS level
- **Failover preparation** - Pre-warm DNS caches with multiple endpoints
- **A/B testing** - Route portions of traffic to different versions
- **Geographic distribution** - Rotate between regional endpoints
- **Maintenance windows** - Gradually shift traffic away from servers under maintenance

## License

MIT License - see [LICENSE](LICENSE) for details.
