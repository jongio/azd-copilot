#!/bin/bash
# Install a PR build of the azd copilot extension
# Usage: ./install-pr.sh <pr_number> <version>
# Example: ./install-pr.sh 123 0.1.0-pr123

set -e

PR_NUMBER="$1"
VERSION="$2"

if [ -z "$PR_NUMBER" ] || [ -z "$VERSION" ]; then
    echo "Usage: $0 <pr_number> <version>"
    echo "Example: $0 123 0.1.0-pr123"
    exit 1
fi

REPO="jongio/azd-copilot"
EXTENSION_ID="jongio.azd.copilot"
TAG="azd-ext-jongio-azd-copilot_${VERSION}"
REGISTRY_URL="https://github.com/$REPO/releases/download/$TAG/pr-registry.json"

echo "🚀 Installing azd copilot PR #$PR_NUMBER (version $VERSION)"
echo ""

# Step 0: Kill any running extension processes
echo "🛑 Stopping any running extension processes..."
pkill -f "jongio-azd-copilot" 2>/dev/null || true
sleep 0.5
echo "   ✓"

# Step 1: Uninstall existing extension
echo "🗑️  Uninstalling existing extension (if any)..."
azd extension uninstall "$EXTENSION_ID" 2>/dev/null || true
rm -rf "$HOME/.azd/extensions/$EXTENSION_ID" 2>/dev/null || true
echo "   ✓"

# Step 2: Download PR registry
echo "📥 Downloading PR registry..."
REGISTRY_PATH="$(pwd)/pr-registry.json"
if ! curl -fsSL "$REGISTRY_URL" -o "$REGISTRY_PATH"; then
    echo "❌ Failed to download registry from $REGISTRY_URL"
    echo "   Make sure the PR build exists and is accessible"
    exit 1
fi
echo "   ✓ Downloaded to: $REGISTRY_PATH"

# Step 3: Add registry source
echo "🔗 Adding PR registry source..."
azd extension source remove "pr-$PR_NUMBER" 2>/dev/null || true
azd extension source add -n "pr-$PR_NUMBER" -t file -l "$REGISTRY_PATH"

# Step 4: Install PR version
echo "📦 Installing version $VERSION..."
rm -rf "$HOME/.azd/cache/"*"$EXTENSION_ID"* 2>/dev/null || true
azd extension install "$EXTENSION_ID" --version "$VERSION"

# Step 5: Verify installation
echo ""
echo "✅ Installation complete!"
echo ""
echo "🔍 Verifying installation..."
if INSTALLED_VERSION=$(azd copilot version 2>&1); then
    echo "   $INSTALLED_VERSION"
    if echo "$INSTALLED_VERSION" | grep -q "$VERSION"; then
        echo ""
        echo "✨ Success! PR build is ready to test."
    else
        echo ""
        echo "⚠️  Version mismatch - expected $VERSION"
    fi
else
    echo "⚠️  Could not verify version"
fi

echo ""
echo "Try these commands:"
echo "  azd copilot version"
echo ""
echo "To restore stable version, run:"
echo "  curl -fsSL https://raw.githubusercontent.com/$REPO/main/scripts/restore-stable.sh | bash"
