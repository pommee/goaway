#!/bin/sh

set -e

target=""
githubUrl="https://github.com"
executable_folder=$(eval echo "~/.local/bin")

get_arch() {
    case $(uname -m) in
        "x86_64" | "amd64" ) echo "amd64" ;;
        "i386" | "i486" | "i586") echo "386" ;;
        "aarch64" | "arm64" | "arm") echo "arm64" ;;
        "mips64el") echo "mips64el" ;;
        "mips64") echo "mips64" ;;
        "mips") echo "mips" ;;
        *) echo "unknown" ;;
    esac
}

get_os() {
    uname -s | tr '[:upper:]' '[:lower:]'
}

get_latest_release() {
    curl --silent "https://api.github.com/repos/pommee/goaway/releases/latest" |
    grep '"tag_name":' |
    sed -E 's/.*"v([^"]+)".*/\1/'
}

main() {
    os=$(get_os)
    arch=$(get_arch)
    version="${1:-$(get_latest_release)}"
    file_name="goaway_${version}_${os}_${arch}.tar.gz"
    downloadFolder="${TMPDIR:-/tmp}"
    downloaded_file="${downloadFolder}/${file_name}"

    echo "[1/3] Downloading ${file_name} to ${downloadFolder}"
    asset_path="https://github.com/pommee/goaway/releases/download/v${version}/${file_name}"

    if ! curl --silent --head --fail "$asset_path" > /dev/null; then
        echo "ERROR: Unable to find a release asset called ${file_name}"
        exit 1
    fi

    echo "Downloading from: ${asset_path}"
    rm -f "${downloaded_file}"
    curl --fail --location --output "${downloaded_file}" "${asset_path}"

    mkdir -p "${executable_folder}"

    echo "[2/3] Installing ${file_name} to ${executable_folder}"
    tar -xzf "${downloaded_file}" -C "${executable_folder}"
    chmod +x "${executable_folder}/goaway"

    echo "[3/3] goaway v${version} was installed successfully to ${executable_folder}"

    echo "Manually add the directory to your \$HOME/.bash_profile (or similar):"
    echo "  export PATH=${executable_folder}:\$PATH"
    exit 0
}

main "$@"
