### Option 1: Docker Installation (Recommended)

**Quick Start:**

```
docker run -d --name goaway \
    -p 53:53/udp \
    -p 53:53/tcp \
    -p 8080:8080 \
    pommee/goaway:latest
```

**Using Docker Compose (Recommended for production):**

Data will **not persist** unless volumes are used!

!!! info "Hashtags are comments"

    Remove them to make use of any setting

```yml
services:
  goaway:
    image: pommee/goaway:latest
    container_name: goaway
    restart: unless-stopped
    # volumes:
    #  - /path/to/config:/app/config  # Custom settings.yaml configuration
    #  - /path/to/data:/app/data      # Database storage location
    environment:
      - DNS_PORT=${DNS_PORT:-53}
      - WEBSITE_PORT=${WEBSITE_PORT:-8080}
    # - DOT_PORT=${DOT_PORT:-853}  # Port for DoT
    # - DOH_PORT=${DOH_PORT:-443}  # Port for DoH
    ports:
      - "${DNS_PORT:-53}:${DNS_PORT:-53}/udp"
      - "${DNS_PORT:-53}:${DNS_PORT:-53}/tcp"
      - "${WEBSITE_PORT:-8080}:${WEBSITE_PORT:-8080}/tcp"
    # - "${DOT_PORT:-853}:${DOT_PORT:-853}/tcp"
    # - "${DOH_PORT:-443}:${DOH_PORT:-443}/tcp"
    cap_add:
      - NET_BIND_SERVICE
      - NET_RAW
```

### Option 2: Quick Install

**Quick Install Script:**

```bash
# Install latest version
curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh

# Install specific version
curl https://raw.githubusercontent.com/pommee/goaway/main/installer.sh | sh /dev/stdin 0.40.4
```

The installer will:

1. Detect your operating system and architecture
2. Download the appropriate binary
3. Install it to `~/.local/bin`
4. Set up necessary permissions

**Manual Installation:**
Download binaries directly from the [releases page](https://github.com/pommee/goaway/releases).

### Option 3: Build from Source

```bash
# Clone the repository
git clone https://github.com/pommee/goaway.git
cd goaway

# Build the frontend
make build

# Build GoAway binary
go build -o goaway

# Start the service
./goaway
```

~~### Option 4: Proxmox~~

!!! warning "Paused"

    Proxmox script support was removed due to the early stage of the project, will likely be back in the future once the first major release is published. Read more [here](https://github.com/pommee/goaway/issues/109).

> ~~If you are planning on running via Proxmox, then there is a helper-script available [here](https://community-scripts.github.io/ProxmoxVE/scripts?id=goaway), created by [Proxmox VE Helper-Scripts (Community Edition)](https://github.com/community-scripts/ProxmoxVE).  
> Great alternative if you don't want to go through an as manual process and prefer not to use Docker.~~

> !!! tip "~~Logs / Credentials~~"
>
> ~~Once the LXC is up and running, logs can be found in `/var/log/goaway.log`. Login credentials can be found either in the log or `~/goaway.creds`.~~
