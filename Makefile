# Makefile for cross-platform video-compress builds

# === Configuration ===
BINARY_NAME = video-compress
VERSION     = $(shell git describe --tags --dirty --always)
DIST_DIR    = dist
GOFLAGS     = -trimpath
LDFLAGS     = -s -w -X main.Version=$(VERSION)
GO          = go

# Target platforms (os_arch)
PLATFORMS = darwin_amd64 linux_amd64 windows_amd64

.PHONY: all tidy build checksum clean

# Default target: tidy, build all, then generate checksums
all: tidy build checksum

# Ensure module dependencies are clean
tidy:
	$(GO) mod tidy

# Build for all defined platforms
build: $(PLATFORMS:%=build-%)

# Platform-specific build rule with .exe for Windows
build-%:
	@echo "Building for $*..."
	@OS=$${*%%_*}; ARCH=$${*#*_}; \
	mkdir -p $(DIST_DIR)/$*; \
	if [ "$$OS" = "windows" ]; then \
		GOOS=$$OS GOARCH=$$ARCH $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" \
			-o $(DIST_DIR)/$*/$(BINARY_NAME).exe ./cmd/compress; \
	else \
		GOOS=$$OS GOARCH=$$ARCH $(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" \
			-o $(DIST_DIR)/$*/$(BINARY_NAME) ./cmd/compress; \
	fi

# Generate SHA256 checksums for all built artifacts
checksum: build
	@echo "Generating SHA256SUMS.txt in $(DIST_DIR)/..."
	@cd $(DIST_DIR) && rm -f SHA256SUMS.txt && \
	for dir in $(PLATFORMS); do \
		for bin in $$dir/*; do \
			sha256sum "$$bin" >> SHA256SUMS.txt; \
		done; \
	done

# Clean build artifacts
clean:
	rm -rf $(DIST_DIR)
