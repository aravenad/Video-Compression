<#
.SYNOPSIS
  Cross-compile video-compress for darwin, linux, windows.

.DESCRIPTION
  Uses $PLATFORMS to drive GOOS/GOARCH settings, emits binaries (with .exe for Windows)
  into dist\<platform>.

.EXAMPLE
  PS> .\build.ps1
  PS> .\build.ps1 -Clean  # Only clean the dist directory
#>

param(
  [string] $Version = $(git describe --tags --dirty --always),
  [string] $DistDir = 'dist',
  [switch] $Clean
)

# Define your target triples
$PLATFORMS = 'darwin_amd64','linux_amd64','windows_amd64'

# Function to clean the distribution directory
function Clear-DistDirectory {
  if (Test-Path $DistDir) { 
    Write-Host "Cleaning $DistDir directory..."
    Remove-Item -Recurse -Force $DistDir 
  }
}

# Clear out old artifacts
Clear-DistDirectory

# Exit if only cleaning was requested
if ($Clean) {
  Write-Host "Clean completed."
  exit 0
}

# Create the distribution directory
New-Item -ItemType Directory -Path $DistDir | Out-Null

# Ensure config directory exists
$configDir = "config"
if (-not (Test-Path $configDir)) {
  Write-Host "Creating config directory..."
  New-Item -ItemType Directory -Force -Path $configDir | Out-Null
  
  # Create default.yaml if it doesn't exist
  $defaultConfig = Join-Path $configDir "default.yaml"
  if (-not (Test-Path $defaultConfig)) {
    @"
# Default configuration for video-compress
output:
  format: mp4
  quality: high

compression:
  preset: medium
  crf: 23
"@ | Set-Content $defaultConfig
    Write-Host "Created default config file: $defaultConfig"
  }
}

foreach ($plat in $PLATFORMS) {
  $parts = $plat.Split('_')
  $env:GOOS   = $parts[0]
  $env:GOARCH = $parts[1]

  $outDir = Join-Path $DistDir $plat
  New-Item -ItemType Directory -Force -Path $outDir | Out-Null

  $exe = if ($env:GOOS -eq 'windows') { '.exe' } else { '' }
  $outFile = Join-Path $outDir ("video-compress$exe")

  Write-Host "Building for $env:GOOS/$env:GOARCH -> $outFile"
  go build -trimpath -ldflags "-s -w -X main.Version=$Version" -o $outFile ./cmd/compress
  
  # Copy config directory to platform output directory
  $platformConfigDir = Join-Path $outDir "config"
  Write-Host "Copying config directory to $platformConfigDir"
  Copy-Item -Path $configDir -Destination $outDir -Recurse -Force
}
