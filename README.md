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
| Variable | Default | Info |
| --------------- | ---------- | ------------------------------------------------------------ |
| GOAWAY_PORT | 53 | Port used for the DNS server. |
| WEBSITE_PORT | 8080 | Port used for the API server. Also serves the website pages. |
| GOAWAY_PASSWORD | No default | Password used for authenticating at the admin dashboard. |

### TODO

| Title          | Description                                                                                                                                                                     | Done          |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------- |
| Login page     | Upon first visit, a login page should be presented. Requiring a login name and password in order to proceed.                                                                    | ✅            |
| Authentication | Provide some form of token for future authentication after login process is complete.                                                                                           | ✅            |
| Logs page      | An overview of all domain lookups that occurs with detailed information. Also add the ability to clear logs.                                                                    | ❌ (progress) |
| Clients page   | Used to see all clients which makes use of the DNS server. Also add settings specific to a certain user, such as...<br>_ Ignore blocked domains.<br>_ Blacklist user.<br>\* ... | ❌ (progress) |
| Settings page  | Add toggles for settings for the server and it's behaviour.                                                                                                                     | ❌ (progress) |
| Server page    | More verbose statistics of the DNS server and web server.                                                                                                                       | ❌            |
| Upstream page  | Ability to add, remove upstreams. And set one as the preferred one to use.                                                                                                      | ❌ (progress) |
