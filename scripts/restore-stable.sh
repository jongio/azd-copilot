#!/bin/bash
# Restore the stable version of the azd copilot extension
# Usage: ./restore-stable.sh

set -e

REPO="jongio/azd-copilot"
EXTENSION_ID="jongio.azd.copilot"

echo "🔄 Restoring stable azd copilot extension"
echo ""

# Kill any running extension processes
echo "🛑 Stopping any running extension processes..."
pkill -f "jongio-azd-copilot" 2>/dev/null || true
pkill -f "copilot" 2>/dev/null || true
sleep 0.5
echo "   ✓"

# Uninstall existing extension
echo "🗑️  Uninstalling existing extension..."
azd extension uninstall "$EXTENSION_ID" 2>/dev/null || true
rm -rf "$HOME/.azd/extensions/$EXTENSION_ID" 2>/dev/null || true
echo "   ✓"

# Remove any PR registry sources
echo "🔗 Removing PR registry sources..."
azd extension source list --output json 2>/dev/null | \
    jq -r '.[] | select(.name | test("^pr-[0-9]+$")) | .name' | \
    while read source; do
        azd extension source remove "$source" 2>/dev/null || true
    done
echo "   ✓"

# Add stable registry source
echo "📋 Adding stable registry source..."
azd extension source remove "$REPO" 2>/dev/null || true
azd extension source add -n "$REPO" -t url -l "https://raw.githubusercontent.com/$REPO/main/registry.json"
echo "   ✓"

# Install stable version
echo "📦 Installing stable version..."
if ! azd extension install "$EXTENSION_ID" --source "$REPO"; then
    echo "❌ Failed to install stable extension"
    exit 1
fi

# Verify installation
echo ""
echo "✅ Stable version restored!"
echo ""
if INSTALLED_VERSION=$(azd copilot version 2>&1); then
    echo "   $INSTALLED_VERSION"
fi
