# Video Compressor

A fast, single-binary Go CLI for cross-platform video compression using FFmpeg presets.

---

## Features

- **Efficient**: Leverages Go’s concurrency for parallel compression jobs.
- **Configurable**: Manage named presets in a YAML config (default `config/default.yaml`).
- **Overrides**: Override codec, ffmpeg preset, or CRF on the fly (`--video-codec`, `--ffpreset`, `--crf`).
- **Multi-platform**: Builds for Linux, macOS, and Windows with a single Makefile or PowerShell script.
- **Queue**: Efficient task queue with configurable worker count (`--jobs`).
- **Packaging**: `dist/` artifacts with SHA256 checksums for reproducible releases.

---

## Installation

### From source (Unix)

1. Clone the repo:
   ```bash
   git clone https://github.com/yourorg/video-compressor.git
   cd video-compressor
   ```
2. Ensure Go 1.24.2 is installed.
3. Build for all platforms:
   ```bash
   make        # requires GNU Make
   ```
4. Binaries and `SHA256SUMS.txt` appear in `dist/`.

### From source (Windows PowerShell)

1. Clone and cd into project.
2. Run the PowerShell script:
   ```powershell
   .\build.ps1   # builds dist\windows_amd64\video-compress.exe
   ```

### Download pre-built binaries

Check the [Releases](https://github.com/yourorg/video-compressor/releases) page for `dist/` archives and `SHA256SUMS.txt`. Verify with:

```bash
cd dist
sha256sum --check SHA256SUMS.txt
```

---

## Quickstart

### List available presets

```bash
video-compress presets list
# → default
```

### Add or update a preset

```bash
video-compress presets add fast \
  --video-codec h264 \
  --preset fast \
  --crf 20
```

### Remove a preset

```bash
video-compress presets remove fast
```

### Compress one or more videos

```bash
video-compress compress [flags] <file1> [file2…]
```

**Flags**:

- `-c`, `--config` `<dir>`: Config directory (default `config/`).
- `-p`, `--preset` `<name>`: Named preset to use (default `default`).
- `--video-codec` `<codec>`: Override preset’s video codec.
- `--ffpreset` `<preset>`: Override preset’s ffmpeg preset.
- `--crf` `<value>`: Override preset’s CRF.
- `-j`, `--jobs` `<n>`: Number of parallel workers (default CPU count).

**Example**:

```bash
video-compress compress -j 4 -o outdir/ --preset fast video1.mp4 video2.mov
```

Output will be in `outdir/` (or alongside inputs, appending `-compressed`).

---

## Benchmarking

A benchmark harness lives in `internal/compressor/compressor_bench_test.go`. Run:

```bash
go test ./internal/compressor -bench=. -benchtime=1s
```

---

## Packaging & Releases

See [docs/packaging.md](docs/packaging.md) for:

- Build instructions
- Checksums & verification
- CI / GitHub Actions workflow

---

## License

[MIT License](LICENSE)
