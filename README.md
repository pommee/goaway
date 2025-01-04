# goaway - DNS sinkhole

![goaway Preview](./resources/dashboard.png)

**[More preview images](./resources/)**

## Acknowledgments

Heavily inspired by [pi-hole](https://github.com/pi-hole/pi-hole).

## 📦 Installation

**goaway** supports the following platforms:

- **Operating Systems**: Linux, macOS, and Windows
- **Architectures**: amd64, arm64, and 386

> [!NOTE]
> Testing has primarily been conducted on **Linux (amd64)**.
> Functionality on macOS and Windows may vary and is not guaranteed.

### Install the Latest Version

To install the latest version of goaway, run the following command:

```shell
curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin
```

This will install the binary specific to your platform.
The binary is placed in `~/.local/bin`.
If the [installer.sh](https://github.com/pommee/goaway/blob/main/installer.sh) script fails, then binaries can be manually downloaded from [releases](https://github.com/pommee/goaway/releases).

## 🛠 Usage

### Starting the Application

To start the servers (dns & web), simply run the following command in your terminal:

```console
$ goaway
```

> Will display some information once the server is started.

![started](./resources/started.png)

```console
$ goaway --help

Flags:
      --disablelogging      If true, then no logs will appear in the container
      --dnsport int         Port for the DNS server (default 53)
  -h, --help                help for goaway
      --loglevel int        0 = DEBUG | 1 = INFO | 2 = WARNING | 3 = ERROR (default 1)
      --noauth              If true, then no authentication is required for the admin dashboard
      --webserverport int   Port for the web server (default 8080)
```

This readme might not always be up to date on the available commands.  
Use `--help` to see what is available.

### Development

Environment variables are used for configuration.
| Variable        | Default    | Info                                                         |
| --------------- | ---------- | ------------------------------------------------------------ |
| GOAWAY_PORT     | 53         | Port used for the DNS server.                                |
| WEBSITE_PORT    | 8080       | Port used for the API server. Also serves the website pages. |
| GOAWAY_PASSWORD | No default | Password used for authenticating at the admin dashboard.     |
