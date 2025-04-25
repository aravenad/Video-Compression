# testdata

This directory contains test assets required for performance benchmarking.

## Prerequisites

- Add a short video file named `sample.mp4` in this folder.
- Recommended duration: 5–15 seconds.
- Recommended file size: ≤ 10 MB.
- Supported format: H.264-encoded MP4.

## Usage

Run the benchmark test for the compressor:

```bash
go test -bench=. -run=^$ ./internal/compressor
```

The `compressor_bench_test.go` will automatically detect `sample.mp4`.

## File Naming

- Filename must be exactly `sample.mp4`.
- Place it at the root of this `testdata` directory.

## Troubleshooting

- Ensure the file path is correct: `.../internal/compressor/testdata/sample.mp4`.
- Verify read permissions on the video file.
- For large videos, consider trimming to meet size/duration recommendations.
