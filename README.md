# goaway - DNS request blocker

![goaway Preview](./resources/preview.png)

## 🙏 Acknowledgments

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

### Development

Environment variables are used for configuration.
| Variable         | Default          | Info                                                                                             |
| ---------------- | ---------------- | ------------------------------------------------------------------------------------------------ |
| GOAWAY_PORT      | 53               | Port used for the DNS server.                                                                    |
| WEBSITE_PORT     | 8080             | Port used for the API server. Also serves the website pages.                                     |
| UPSTREAM_DNS     | 8.8.8.8:53       | IP and port the server uses to resolve domain names.                                             |
| BLACKLIST_PATH   | ./blacklist.json | File containing all domain names that will be blocked.                                           |
| COUNTER_PATH     | ./counter.json   | Keeps track of of various statistics.                                                            |
| REQUEST_LOG_FILE | ./requests.json  | Storage for all requests served. Contains timestamps, domain names and if a request was blocked. |

