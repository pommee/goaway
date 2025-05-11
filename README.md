# GoAway - DNS Sinkhole

![GitHub Release](https://img.shields.io/github/v/release/pommee/goaway)
![GitHub Actions Workflow Status](https://img.shields.io/github/actions/workflow/status/pommee/goaway/release.yml)
![GitHub Downloads (all assets, all releases)](https://img.shields.io/github/downloads/pommee/goaway/total?color=cornflowerblue)

A lightweight DNS sinkhole for blocking unwanted domains, inspired by Pi-hole.

![goaway Dashboard Preview](./resources/dashboard.png)

**[View more screenshots](./resources/PREVIEW.md)**

## 🌟 Features

- DNS-level domain blocking
- Web-based admin dashboard
- Cross-platform support
- Docker support
- Customizable blocking rules
- Real-time statistics
- Low resource footprint
- And much more...

## 📋 Requirements

- For binary installation: Linux, macOS, or Windows
- For Docker installation: Docker and Docker Compose
- Supported architectures: amd64, arm64, and 386

## 📦 Installation

### Option 1: Docker Installation (Recommended)

Run GoAway in a containerized environment:

```shell
docker run pommee/goaway:latest

# Best is to use the compose file in this repository
docker compose up -d
```

Use compose for more customization, example can be found [here](https://github.com/pommee/goaway/blob/main/docker-compose.yml)

### Option 2: Quick Install

Install using the installation script:

```bash
# Latest version available
curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin

# Specific version
curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin 0.40.4
```

The installer will:

1. Detect your operating system and architecture
2. Download the appropriate binary
3. Install it to `~/.local/bin`
4. Set up necessary permissions

If the installer fails, you can manually download binaries from the [releases page](https://github.com/pommee/goaway/releases).

### Option 3: Build from source

Last option is to build from source:

```bash
# Build the frontend client that will be embedded into the binary
make build

# Build GoAway binary
go build -o goaway

# Start
./goaway
```

## 🚀 Getting Started

### Basic Usage

Start the DNS and web server with default settings:

```bash
goaway
```

You'll see a startup message confirming the services are running:

![Startup Screen](./resources/started.png)

### Configuration Options

```bash
goaway --help

GoAway is a DNS sinkhole with a web interface

Usage:
  goaway [flags]

Flags:
      --auth                       Disable authentication for admin dashboard (default true)
      --dns-port int               Port for the DNS server (default 53)
  -h, --help                       help for goaway
      --log-level int              0 = DEBUG | 1 = INFO | 2 = WARNING | 3 = ERROR (default 1)
      --logging                    Toggle logging (default true)
      --statistics-retention int   Days to keep statistics (default 1)
      --webserver-port int         Port for the web server (default 8080)
```

### Custom Configuration

The default settings are defined in `settings.json`. You can customize it by modifying the values as needed.

## ⚠️ Platform Support

| Platform | Architecture | Support Level |
| -------- | ------------ | ------------- |
| Linux    | amd64        | Full          |
| Linux    | arm64        | Full          |
| Linux    | 386          | Full          |
| macOS    | amd64        | Beta          |
| macOS    | arm64        | Beta          |
| Windows  | amd64        | Beta          |
| Windows  | 386          | Beta          |

> **Note**: Primary testing is conducted on Linux (amd64). While the aim is to support all listed platforms, functionality on macOS and Windows may vary.

## Dev

The dashboard and servers are started separately, reason being hot-reloads and not having to embed the client into the binary.

```bash
make dev-website
make dev-server
```

## 💫 Contributing

You can start a new [discussion here](https://github.com/pommee/goaway/discussions) if features are wanted.
Please report any issues encountered [here](https://github.com/pommee/goaway/issues/new).

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

This project is heavily inspired by [Pi-hole](https://github.com/pi-hole/pi-hole). Thanks to all people involved for their work.
