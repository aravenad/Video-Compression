# Packaging & Releases

This document describes how to build, test, verify, and release video-compress binaries.

## Prerequisites

- Go 1.24.2 (pinned in `go.mod`)
- GNU Make (or WSL / Linux / macOS) **OR** PowerShell (for Windows users)
- `ffmpeg` must be installed on your `PATH` for compression tests/benchmarks.

## Local Development

From the **project root** (where `Makefile` and `go.mod` live):

1. **Tidy modules**

   ```bash
   # Using Make
   make tidy

   # Using PowerShell
   go mod tidy
   ```

2. **Run tests**

   ```bash
   # Using Make/Bash or PowerShell
   go test ./... -timeout 2m
   ```

3. **Run benchmarks**

   ```bash
   # Using Make/Bash or PowerShell
   go test ./internal/compressor -bench=BenchmarkCompressDefault -benchtime=1s
   ```

4. **Build for all platforms and generate checksums**

   ```bash
   # Using Make
   make

   # Using PowerShell
   ./build.ps1
   ./checksum.ps1
   ```

   Artifacts and `SHA256SUMS.txt` will be in:

   - `dist/darwin_amd64/`
   - `dist/linux_amd64/`
   - `dist/windows_amd64/`

5. **Clean artifacts**

   ```bash
   # Using Make
   make clean

   # Using PowerShell
   ./build.ps1 -Clean
   ```

## Verifying Downloads

After downloading your platform-specific binary and the accompanying `SHA256SUMS.txt`:

```bash
# Verify the SHA256 sums (from the dist/ directory)
sha256sum --check SHA256SUMS.txt
```

If provided, verify the Git tag's GPG signature for authenticity:

```bash
# Fetch tags and verify signature
git fetch --tags && git tag -v v1.0.0
```

## CI & Release

We use GitHub Actions to automate:

- **CI** on PRs and `main` pushes: tests, benchmarks, cross-platform builds, and checksum generation.
- **Release** on version tags (`vX.Y.Z`): builds artifacts, generates checksums, and publishes a GitHub Release with binaries and `SHA256SUMS.txt`.

See [`.github/workflows/ci-and-release.yml`](.github/workflows/ci-and-release.yml) for details.

### Releasing a new version

1. **Bump your version tag:**
   ```bash
   git tag v1.0.0
   git push origin v1.0.0
   ```
2. **Actions** will automatically build, checksum, and publish the release.

## Reproducible Builds

- Go version pinned to `1.24.2` in `go.mod`.
- `GOFLAGS=-trimpath` and LDFLAGS `-s -w` strip paths and debug symbols.
- Binaries and checksums in `dist/` are deterministic given the same tag and source.
