# goaway - DNS Sinkhole

A lightweight DNS sinkhole for blocking unwanted domains, inspired by Pi-hole.

![goaway Dashboard Preview](./resources/dashboard.png)

**[View more screenshots](./resources/)**

## 🌟 Features

- DNS-level domain blocking
- Web-based admin dashboard
- Cross-platform support
- Docker support
- Customizable blocking rules
- Real-time statistics
- Low resource footprint

## 📋 Requirements

- For binary installation: Linux, macOS, or Windows
- For Docker installation: Docker and Docker Compose
- Supported architectures: amd64, arm64, and 386

## 📦 Installation

### Option 1: Quick Install (Recommended)

Install the latest version with the installation script:

```bash
curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin
```

The installer will:
1. Detect your operating system and architecture
2. Download the appropriate binary
3. Install it to `~/.local/bin`
4. Set up necessary permissions

If the installer fails, you can manually download binaries from the [releases page](https://github.com/pommee/goaway/releases).

### Option 2: Docker Installation

Run goaway in a containerized environment. This requires a clone of this repository and running the following in project root:

```bash
make start
```

## 🚀 Getting Started

### Basic Usage

Start the DNS and web servers with default settings:

```bash
goaway
```

You'll see a startup message confirming the services are running:

![Startup Screen](./resources/started.png)

### Configuration Options

```bash
goaway --help

Available flags:
      --disablelogging      Disable all logging output
      --dnsport int         DNS server port (default: 53)
  -h, --help               Show help message
      --loglevel int       Set logging level:
                            0 = DEBUG
                            1 = INFO (default)
                            2 = WARNING
                            3 = ERROR
      --noauth             Disable dashboard authentication
      --webserverport int  Web server port (default: 8080)
```

### Custom Configuration

The default settings are defined in `settings.json`. You can customize it by odifying the values as needed.

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

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

This project is heavily inspired by [Pi-hole](https://github.com/pi-hole/pi-hole). Thanks to all people involved for their work.
