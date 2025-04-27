<#
.SYNOPSIS
  Generate or verify SHA256SUMS.txt in dist/
#>

param(
  [switch] $Verify
)

$DistDir = 'dist'
$SumFile = Join-Path $DistDir 'SHA256SUMS.txt'

if (-not $Verify) {
  Remove-Item $SumFile -ErrorAction SilentlyContinue
  Get-ChildItem "$DistDir\*\*" -File | ForEach-Object {
    $rel = $_.FullName.Substring((Get-Location).Path.Length+1)
    $hash = Get-FileHash $_.FullName -Algorithm SHA256
    "{0} *{1}" -f $hash.Hash.ToLower(), $rel
  } | Set-Content $SumFile
  Write-Host "Wrote checksums to $SumFile"
} else {
  # Verify mode
  Get-Content $SumFile | ForEach-Object {
    $parts = $_ -split ' \*'
    $expected = $parts[0]
    $path = Join-Path (Get-Location) $parts[1]
    $actual = (Get-FileHash $path -Algorithm SHA256).Hash.ToLower()
    if ($actual -ne $expected) {
      Write-Error "$parts[1]: checksum mismatch"
    } else {
      Write-Host "$parts[1]: OK"
    }
  }
}
