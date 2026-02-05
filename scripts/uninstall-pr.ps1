#!/usr/bin/env pwsh
<#
.SYNOPSIS
    Uninstall a PR build of the azd copilot extension
.PARAMETER PrNumber
    The PR number to uninstall
.EXAMPLE
    .\uninstall-pr.ps1 -PrNumber 123
#>

param(
    [Parameter(Mandatory=$true)]
    [int]$PrNumber
)

$ErrorActionPreference = 'Stop'
$extensionId = "jongio.azd.copilot"

Write-Host "🗑️  Uninstalling azd copilot PR #$PrNumber" -ForegroundColor Cyan
Write-Host ""

# Kill any running extension processes
Write-Host "🛑 Stopping any running extension processes..." -ForegroundColor Gray
$processNames = @("jongio-azd-copilot-windows-amd64", "jongio-azd-copilot-windows-arm64")
foreach ($name in $processNames) {
    Get-Process -Name $name -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
}
Start-Sleep -Milliseconds 500
Write-Host "   ✓" -ForegroundColor DarkGray

# Uninstall extension
Write-Host "📦 Uninstalling extension..." -ForegroundColor Gray
azd extension uninstall $extensionId 2>&1 | Out-Null

$extensionDir = Join-Path $env:USERPROFILE ".azd\extensions\$extensionId"
if (Test-Path $extensionDir) {
    Remove-Item -Path $extensionDir -Recurse -Force -ErrorAction SilentlyContinue
}
Write-Host "   ✓" -ForegroundColor DarkGray

# Remove PR registry source
Write-Host "🔗 Removing PR registry source..." -ForegroundColor Gray
azd extension source remove "pr-$PrNumber" 2>$null
Write-Host "   ✓" -ForegroundColor DarkGray

# Clean up registry file
$registryPath = Join-Path $PWD "pr-registry.json"
if (Test-Path $registryPath) {
    Remove-Item -Path $registryPath -Force -ErrorAction SilentlyContinue
}

Write-Host ""
Write-Host "✅ PR build uninstalled!" -ForegroundColor Green
