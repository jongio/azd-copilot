#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Restore the stable version of the azd copilot extension
.EXAMPLE
    .\restore-stable.ps1
#>

$ErrorActionPreference = 'Stop'
$repo = "jongio/azd-copilot"
$extensionId = "jongio.azd.copilot"
$sourceName = "jongio"
$registryUrl = "https://jongio.github.io/azd-extensions/registry.json"

Write-Host "🔄 Restoring stable azd copilot extension" -ForegroundColor Cyan
Write-Host ""

# Kill any running extension processes
Write-Host "🛑 Stopping any running extension processes..." -ForegroundColor Gray
$processNames = @("jongio-azd-copilot-windows-amd64", "jongio-azd-copilot-windows-arm64")
foreach ($name in $processNames) {
    Get-Process -Name $name -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
}
Start-Sleep -Milliseconds 500
Write-Host "   ✓" -ForegroundColor DarkGray

# Uninstall existing extension
Write-Host "🗑️  Uninstalling existing extension..." -ForegroundColor Gray
azd extension uninstall $extensionId 2>&1 | Out-Null

$extensionDir = Join-Path $env:USERPROFILE ".azd\extensions\$extensionId"
if (Test-Path $extensionDir) {
    Remove-Item -Path $extensionDir -Recurse -Force -ErrorAction SilentlyContinue
}

# Remove any PR registry sources
Write-Host "🔗 Removing PR registry sources..." -ForegroundColor Gray
$sources = azd extension source list --output json 2>$null | ConvertFrom-Json
foreach ($source in $sources) {
    if ($source.name -match "^pr-\d+$") {
        azd extension source remove $source.name 2>$null
    }
}
Write-Host "   ✓" -ForegroundColor DarkGray

# Add stable registry source
Write-Host "📋 Adding stable registry source..." -ForegroundColor Gray
azd extension source remove $sourceName 2>$null
azd extension source add -n $sourceName -t url -l $registryUrl
Write-Host "   ✓" -ForegroundColor DarkGray

# Install stable version
Write-Host "📦 Installing stable version..." -ForegroundColor Gray
azd extension install $extensionId --source $sourceName
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to install stable extension" -ForegroundColor Red
    exit 1
}

# Verify installation
Write-Host ""
Write-Host "✅ Stable version restored!" -ForegroundColor Green
Write-Host ""
$installedVersion = azd copilot version 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "   $installedVersion" -ForegroundColor DarkGray
}
