#!/bin/bash

BINARY_NAME="goaway"
USERNAME=$(whoami)
INSTALL_DIR="/home/$USERNAME"
TMP_DIR=$(mktemp -d)
ORIGINAL_CMD=$(ps -o args= -C $BINARY_NAME | head -n 1)
LATEST_RELEASE=$(curl -s "https://api.github.com/repos/pommee/goaway/releases/latest")
LATEST_VERSION=$(echo "$LATEST_RELEASE" | jq -r '.tag_name')
ASSET_URL=$(echo "$LATEST_RELEASE" | jq -r '.assets[] | select(.name | endswith("linux_amd64.tar.gz")) | .browser_download_url')

if [[ -z "$ASSET_URL" ]]; then
    echo "[Error] No valid asset URL found in the GitHub release."
    exit 1
fi

echo "[INFO] Downloading $BINARY_NAME version $LATEST_VERSION..."
curl -L -o "$TMP_DIR/$BINARY_NAME.tar.gz" "$ASSET_URL"

echo "[INFO] Extracting $BINARY_NAME..."
tar -xzf "$TMP_DIR/$BINARY_NAME.tar.gz" -C "$TMP_DIR"

echo "[INFO] Stopping $BINARY_NAME..."
pkill $BINARY_NAME

echo "[INFO] Moving binary $TMP_DIR/$BINARY_NAME"
mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

echo "[INFO] Starting goaway again..."
exec $ORIGINAL_CMD

rm -rf "$TMP_DIR"

echo "Update completed successfully."
