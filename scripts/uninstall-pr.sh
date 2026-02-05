#!/bin/bash
# Uninstall a PR build of the azd copilot extension
# Usage: ./uninstall-pr.sh <pr_number>

set -e

PR_NUMBER="$1"

if [ -z "$PR_NUMBER" ]; then
    echo "Usage: $0 <pr_number>"
    exit 1
fi

EXTENSION_ID="jongio.azd.copilot"

echo "🗑️  Uninstalling azd copilot PR #$PR_NUMBER"
echo ""

# Kill any running extension processes
echo "🛑 Stopping any running extension processes..."
pkill -f "jongio-azd-copilot" 2>/dev/null || true
sleep 0.5
echo "   ✓"

# Uninstall extension
echo "📦 Uninstalling extension..."
azd extension uninstall "$EXTENSION_ID" 2>/dev/null || true
rm -rf "$HOME/.azd/extensions/$EXTENSION_ID" 2>/dev/null || true
echo "   ✓"

# Remove PR registry source
echo "🔗 Removing PR registry source..."
azd extension source remove "pr-$PR_NUMBER" 2>/dev/null || true
echo "   ✓"

# Clean up registry file
rm -f "$(pwd)/pr-registry.json" 2>/dev/null || true

echo ""
echo "✅ PR build uninstalled!"
