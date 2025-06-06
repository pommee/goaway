#!/bin/bash

BINARY_PATH=${1:-$(pwd)/goaway}
BINARY_NAME=$(basename "$BINARY_PATH")
TMP_DIR=$(mktemp -d)
ORIGINAL_CMD=$(ps -o args= -C "$BINARY_NAME" | head -n 1)

LATEST_RELEASE=$(curl -s "https://api.github.com/repos/pommee/goaway/releases/latest")
if [[ $? -ne 0 ]] || [[ -z "$LATEST_RELEASE" ]]; then
    echo "[error] Failed to fetch release information."
    rm -rf "$TMP_DIR"
    exit 1
fi

LATEST_VERSION=$(echo "$LATEST_RELEASE" | jq -r '.tag_name')
ASSET_URL=$(echo "$LATEST_RELEASE" | jq -r '.assets[] | select(.name | endswith("linux_amd64.tar.gz")) | .browser_download_url')

if [[ -z "$ASSET_URL" ]] || [[ "$ASSET_URL" == "null" ]]; then
    echo "[error] No valid asset URL found in the GitHub release."
    rm -rf "$TMP_DIR"
    exit 1
fi

echo "[info] Found version $LATEST_VERSION"
echo "[info] Downloading asset $ASSET_URL"
if ! curl -L --progress-bar -o "$TMP_DIR/$BINARY_NAME.tar.gz" "$ASSET_URL"; then
    echo "[error] Failed to download the binary."
    rm -rf "$TMP_DIR"
    exit 1
fi

echo "[info] Extracting $BINARY_NAME"
if ! tar -xzf "$TMP_DIR/$BINARY_NAME.tar.gz" -C "$TMP_DIR"; then
    echo "[error] Failed to extract the binary."
    rm -rf "$TMP_DIR"
    exit 1
fi

if [[ ! -f "$TMP_DIR/$BINARY_NAME" ]]; then
    echo "[error] Extracted binary not found."
    rm -rf "$TMP_DIR"
    exit 1
fi

if [[ -f "$BINARY_PATH" ]]; then
    cp "$BINARY_PATH" "$BINARY_PATH.backup"
    echo "[info] Backed up original binary to $BINARY_PATH.backup"
fi

echo "[info] Stopping $BINARY_NAME..."
pkill "$BINARY_NAME"

echo "[info] Moving binary from $TMP_DIR/$BINARY_NAME to $BINARY_PATH"
if ! mv "$TMP_DIR/$BINARY_NAME" "$BINARY_PATH"; then
    echo "[error] Failed to move binary."
    if [[ -f "$BINARY_PATH.backup" ]]; then
        mv "$BINARY_PATH.backup" "$BINARY_PATH"
        echo "[info] Restored original binary."
    fi
    rm -rf "$TMP_DIR"
    exit 1
fi

chmod +x "$BINARY_PATH"

rm -rf "$TMP_DIR"

echo "[info] Update completed successfully."

if [[ -n "$ORIGINAL_CMD" ]]; then
    echo "[info] Starting $BINARY_NAME again..."
    exec $ORIGINAL_CMD
else
    echo "[info] No original command found. Please start $BINARY_NAME manually."
fi
