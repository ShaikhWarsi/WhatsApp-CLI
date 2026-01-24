#!/bin/bash

echo "Downloading WhatsApp CLI by Parth Bhanti..."

# Detect OS
OS="$(uname -s)"
ARCH="$(uname -m)"

case "$OS" in
    Linux)     BINARY="whatsapp-linux-amd64" ;;
    Darwin)    
        if [ "$ARCH" = "arm64" ]; then
            BINARY="whatsapp-mac-arm"
        else
            BINARY="whatsapp-mac-intel"
        fi
        ;;
    *)         echo "Unsupported OS"; exit 1 ;;
esac

URL="https://github.com/parthbhanti22/WhatsApp-CLI/releases/latest/download/$BINARY"
curl -L -o whatsapp-cli "$URL"

# Make executable and move to bin
chmod +x whatsapp-cli
sudo mv whatsapp-cli /usr/local/bin/whatsapp-cli

echo "Success! Type 'whatsapp-cli' to run."