# Video Compressor

A fast, single-binary Go CLI for cross-platform video compression using FFmpeg presets.

---

## Features

- **Efficient**: Leverages Go's concurrency for parallel compression jobs.
- **Configurable**: Manage named presets with descriptions in a YAML config (default `config/default.yaml`).
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
# → hq
# → master-4k
# → mobile-hevc
# → standard
# → web-fast
# → webm-vp9
```

### Show preset details

```bash
video-compress presets show hq
# → Preset: hq
# → Description: High quality preset for important videos with minimal quality loss
# → Video codec: libx264
# → FFmpeg preset: slow
# → CRF value: 18
```

### Add or update a preset

```bash
video-compress presets add fast \
  --video-codec h264 \
  --preset fast \
  --crf 20 \
  --description "Fast encoding with good quality"
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

- `-c`, `--config` `<dir>`: Config directory (default `config/`).
- `-p`, `--preset` `<name>`: Named preset to use (default `default`).
- `--video-codec` `<codec>`: Override preset's video codec.
- `--ffpreset` `<preset>`: Override preset's ffmpeg preset.
- `--crf` `<value>`: Override preset's CRF.
- `-j`, `--jobs` `<n>`: Number of parallel workers (default CPU count).

**Example**:

```bash
video-compress compress -j 4 --preset hq video1.mp4
```

Output will be alongside inputs, appending `-compressed`.  
**If you specify an output directory (with `--output`), it will be created automatically if it does not exist.**

---

## Built-in Presets

The application comes with several pre-configured presets:

| Preset      | Description                                                              | Video Codec | CRF | FFmpeg Preset |
| ----------- | ------------------------------------------------------------------------ | ----------- | --- | ------------- |
| default     | General purpose preset with good balance of quality and file size        | libx264     | 23  | medium        |
| hq          | High quality preset for important videos with minimal quality loss       | libx264     | 18  | slow          |
| master-4k   | Ultra high quality for master copies and 4K content archiving            | libx264     | 12  | veryslow      |
| mobile-hevc | HEVC/H.265 compression optimized for mobile devices and small file sizes | libx265     | 28  | fast          |
| standard    | Standard H.264 compression suitable for most purposes                    | libx264     | 23  | medium        |
| web-fast    | Fast encoding with smaller file sizes for web sharing and streaming      | libx264     | 28  | fast          |
| webm-vp9    | VP9 codec for web publishing on platforms that prefer WebM format        | libvpx-vp9  | 30  | medium        |

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
