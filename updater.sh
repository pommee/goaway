#!/bin/bash

REPO="pommee/goaway"
BINARY_NAME="goaway"
INSTALL_DIR="/home/$USER/dev/goaway"
TMP_DIR=$(mktemp -d)
ORIGINAL_CMD="$GOAWAY_CMD"

if [[ -z "$ORIGINAL_CMD" ]]; then
    echo "Error: Could not determine the original command for $BINARY_NAME."
    exit 1
fi

LATEST_RELEASE=$(curl -s "https://api.github.com/repos/$REPO/releases/latest")

if ! echo "$LATEST_RELEASE" | jq empty > /dev/null 2>&1; then
    echo "Error: Invalid JSON response from GitHub API."
    exit 1
fi

LATEST_VERSION=$(echo "$LATEST_RELEASE" | jq -r '.tag_name')
ASSET_URL=$(echo "$LATEST_RELEASE" | jq -r '.assets[] | select(.name | endswith("linux_amd64.tar.gz")) | .browser_download_url')

if [[ -z "$ASSET_URL" ]]; then
    echo "Error: No valid asset URL found in the GitHub release."
    exit 1
fi

CURRENT_VERSION=$($BINARY_NAME --version 2>/dev/null || echo "unknown")
if [[ "$CURRENT_VERSION" == *"$LATEST_VERSION"* ]]; then
    echo "The latest version ($LATEST_VERSION) is already installed."
    exit 0
fi

echo "Downloading $BINARY_NAME version $LATEST_VERSION..."
curl -L -o "$TMP_DIR/$BINARY_NAME.tar.gz" "$ASSET_URL"

if [[ ! -f "$TMP_DIR/$BINARY_NAME.tar.gz" ]]; then
    echo "Error: Failed to download the tarball."
    exit 1
fi

echo "Extracting $BINARY_NAME..."
tar -xzf "$TMP_DIR/$BINARY_NAME.tar.gz" -C "$TMP_DIR"

echo "Stopping $BINARY_NAME..."
pkill $BINARY_NAME

echo "Moving binary $TMP_DIR/$BINARY_NAME"
echo $ORIGINAL_CMD

echo "Installing new version of $BINARY_NAME..."
mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"

echo "Restarting $BINARY_NAME..."

echo "Restarting with command: $ORIGINAL_CMD"
exec $ORIGINAL_CMD

rm -rf "$TMP_DIR"

echo "Update completed successfully."
