# palm installer for Windows
# Usage: irm https://msalah0e.github.io/palm/install.ps1 | iex

$ErrorActionPreference = "Stop"

$repo = "msalah0e/palm"
$installDir = "$env:LOCALAPPDATA\palm"
$binName = "palm.exe"

Write-Host ""
Write-Host "  🌴 palm installer for Windows" -ForegroundColor Green
Write-Host "  ================================" -ForegroundColor DarkGray
Write-Host ""

# Detect architecture
$arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    Write-Host "  Warning: ARM64 detected. Using amd64 build (runs via emulation)." -ForegroundColor Yellow
}

# Get latest release
Write-Host "  Fetching latest release..." -ForegroundColor DarkGray
$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
$asset = $release.assets | Where-Object { $_.name -like "palm-windows-$arch.zip" }

if (-not $asset) {
    Write-Host "  Error: No Windows build found for $version" -ForegroundColor Red
    exit 1
}

Write-Host "  Found: palm $version" -ForegroundColor Green

# Download
$tmpDir = Join-Path $env:TEMP "palm-install"
$zipFile = Join-Path $tmpDir "palm.zip"
New-Item -ItemType Directory -Force -Path $tmpDir | Out-Null

Write-Host "  Downloading $($asset.name)..." -ForegroundColor DarkGray
Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zipFile

# Extract
Write-Host "  Extracting..." -ForegroundColor DarkGray
Expand-Archive -Path $zipFile -DestinationPath $tmpDir -Force

# Install
New-Item -ItemType Directory -Force -Path $installDir | Out-Null
$exeSource = Get-ChildItem -Path $tmpDir -Filter "*.exe" -Recurse | Select-Object -First 1
Copy-Item -Path $exeSource.FullName -Destination (Join-Path $installDir $binName) -Force

# Clean up
Remove-Item -Path $tmpDir -Recurse -Force

# Add to PATH if not already there
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    Write-Host "  Adding to PATH..." -ForegroundColor DarkGray
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    $env:Path = "$env:Path;$installDir"
    Write-Host "  Added $installDir to user PATH" -ForegroundColor Green
}

Write-Host ""
Write-Host "  ✓ palm $version installed to $installDir" -ForegroundColor Green
Write-Host ""
Write-Host "  Get started:" -ForegroundColor DarkGray
Write-Host "    palm --version          # Verify install"
Write-Host "    palm search             # Browse AI tools"
Write-Host "    palm install ollama     # Install a tool"
Write-Host "    palm doctor             # Check your setup"
Write-Host ""
Write-Host "  Shell completions:" -ForegroundColor DarkGray
Write-Host "    palm completion powershell | Out-String | Invoke-Expression"
Write-Host ""

# Verify
& (Join-Path $installDir $binName) --version
