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
function Clean-DistDirectory {
  if (Test-Path $DistDir) { 
    Write-Host "Cleaning $DistDir directory..."
    Remove-Item -Recurse -Force $DistDir 
  }
}

# Clean out old artifacts
Clean-DistDirectory

# Exit if only cleaning was requested
if ($Clean) {
  Write-Host "Clean completed."
  exit 0
}

# Create the distribution directory
New-Item -ItemType Directory -Path $DistDir | Out-Null

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
}
