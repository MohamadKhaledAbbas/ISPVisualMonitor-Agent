#!/bin/bash
# ISP Visual Monitor Agent - Linux Installation Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/ispagent"
LOG_DIR="/var/log/ispagent"
SERVICE_FILE="/etc/systemd/system/ispagent.service"
REPO_URL="https://github.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent"

echo -e "${GREEN}ISP Visual Monitor Agent - Installation${NC}"
echo "========================================"
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then 
    echo -e "${RED}Error: Please run as root (use sudo)${NC}"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
case $ARCH in
    x86_64)
        BINARY_NAME="ispagent-linux-amd64"
        ;;
    aarch64|arm64)
        BINARY_NAME="ispagent-linux-arm64"
        ;;
    armv7l)
        BINARY_NAME="ispagent-linux-armv7"
        ;;
    *)
        echo -e "${RED}Error: Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

echo "Detected architecture: $ARCH"
echo "Binary: $BINARY_NAME"
echo ""

# Get latest release version
echo "Fetching latest release..."
LATEST_VERSION=$(curl -s https://api.github.com/repos/MohamadKhaledAbbas/ISPVisualMonitor-Agent/releases/latest | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST_VERSION" ]; then
    echo -e "${YELLOW}Warning: Could not fetch latest version. Using 'latest'${NC}"
    LATEST_VERSION="latest"
fi

echo "Version: $LATEST_VERSION"
echo ""

# Download binary
echo "Downloading agent binary..."
DOWNLOAD_URL="${REPO_URL}/releases/download/${LATEST_VERSION}/${BINARY_NAME}.tar.gz"
TEMP_DIR=$(mktemp -d)
cd $TEMP_DIR

curl -L -o ispagent.tar.gz "$DOWNLOAD_URL" || {
    echo -e "${RED}Error: Failed to download binary${NC}"
    rm -rf $TEMP_DIR
    exit 1
}

# Extract and install
echo "Installing agent..."
tar -xzf ispagent.tar.gz
chmod +x $BINARY_NAME
mv $BINARY_NAME $INSTALL_DIR/ispagent

# Create directories
echo "Creating directories..."
mkdir -p $CONFIG_DIR
mkdir -p $LOG_DIR

# Download example config
echo "Installing configuration..."
if [ ! -f "$CONFIG_DIR/agent.yaml" ]; then
    curl -s -o $CONFIG_DIR/agent.yaml.example \
        https://raw.githubusercontent.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/main/configs/agent.yaml.example
    cp $CONFIG_DIR/agent.yaml.example $CONFIG_DIR/agent.yaml
    echo -e "${GREEN}Created default configuration: $CONFIG_DIR/agent.yaml${NC}"
else
    echo -e "${YELLOW}Configuration already exists, skipping...${NC}"
fi

# Create user if not exists
echo "Creating system user..."
if ! id -u ispagent >/dev/null 2>&1; then
    useradd -r -s /bin/false -d /var/lib/ispagent ispagent
    echo "Created user: ispagent"
fi

# Set permissions
chown -R ispagent:ispagent $LOG_DIR
chmod 750 $LOG_DIR
chmod 640 $CONFIG_DIR/agent.yaml

# Install systemd service
if command -v systemctl >/dev/null 2>&1; then
    echo "Installing systemd service..."
    curl -s -o $SERVICE_FILE \
        https://raw.githubusercontent.com/MohamadKhaledAbbas/ISPVisualMonitor-Agent/main/deploy/systemd/ispagent.service
    
    systemctl daemon-reload
    systemctl enable ispagent
    
    echo -e "${GREEN}Systemd service installed and enabled${NC}"
fi

# Cleanup
cd /
rm -rf $TEMP_DIR

echo ""
echo -e "${GREEN}Installation complete!${NC}"
echo ""
echo "Next steps:"
echo "1. Edit configuration: $CONFIG_DIR/agent.yaml"
echo "2. Add your license key and router details"
echo "3. Start the agent: systemctl start ispagent"
echo "4. Check status: systemctl status ispagent"
echo "5. View logs: journalctl -u ispagent -f"
echo ""
echo "Documentation: $REPO_URL/tree/main/docs"
echo ""
