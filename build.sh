#!/bin/bash
# Build script for Goku CLI

# Exit immediately if a command exits with a non-zero status.
set -e

# --- Configuration ---
OUTPUT_DIR="bin"                    # Directory to store built binaries
IMAGE_NAME="goku-builder"           # Name for the Docker build image
LDFLAGS_STR="-s -w"                 # Linker flags to strip binary
MAIN_PACKAGE="./cmd/goku/"          # Path to the main Go package

# --- Ensure Output Directory Exists ---
echo "Creating output directory: ${OUTPUT_DIR}"
mkdir -p "${OUTPUT_DIR}"

# --- Build Linux Binary using Docker ---
echo "--- Building Linux Binary (via Docker) ---"

echo "Building Docker image ${IMAGE_NAME}..."
# Build the Docker image using the Dockerfile in the current directory
# Assumes Dockerfile uses a Debian base as per previous recommendation
docker build -t ${IMAGE_NAME} .

echo "Running Docker container to build Linux binary..."
# Run the container to copy the built binary via volume mount
docker run --rm \
  -v "$(pwd)/${OUTPUT_DIR}:/app/bin" \
  ${IMAGE_NAME} \
  echo "Linux binary build complete (occurred during image build), copied via volume mount."

echo "Linux build process finished."
echo "----------------------------------------"


# --- Build macOS Binary Natively (for Host Architecture) ---
# These commands run directly on the host machine (your Mac)
# Requires Go and Xcode Command Line Tools to be installed locally
echo "--- Building macOS Binary (Natively for Host Architecture) ---"

# Check if running on macOS
if [ "$(uname)" == "Darwin" ]; then
    # Detect host architecture
    HOST_ARCH=$(uname -m)
    TARGET_GOARCH=""
    OUTPUT_SUFFIX=""

    # Determine GOARCH and filename suffix based on host architecture
    if [ "${HOST_ARCH}" = "x86_64" ]; then
        TARGET_GOARCH="amd64"
        OUTPUT_SUFFIX="amd64"
        echo "Detected Intel Mac (x86_64), setting GOARCH=amd64"
    elif [ "${HOST_ARCH}" = "arm64" ]; then
        TARGET_GOARCH="arm64"
        OUTPUT_SUFFIX="arm64"
        echo "Detected Apple Silicon Mac (arm64), setting GOARCH=arm64"
    else
        echo "Warning: Unsupported macOS architecture detected: ${HOST_ARCH}. Skipping native macOS build."
        # Skip the build by not setting TARGET_GOARCH
    fi

    # Only proceed if a supported architecture was detected
    if [ -n "${TARGET_GOARCH}" ]; then
        echo "Building for macOS ${TARGET_GOARCH}..."
        # Ensure CGO is enabled
        # Add -v for verbose output, helpful for debugging Cgo issues
        CGO_ENABLED=1 GOOS=darwin GOARCH=${TARGET_GOARCH} go build -v -ldflags="${LDFLAGS_STR}" -o "${OUTPUT_DIR}/goku-darwin-${OUTPUT_SUFFIX}" "${MAIN_PACKAGE}"
        echo "Finished native macOS ${TARGET_GOARCH} build."
    fi
else
    echo "Not running on macOS, skipping native macOS build."
fi
echo "----------------------------------------"


# --- Build Windows Binary (Placeholder) ---
# CGO cross-compilation from macOS to Windows is non-trivial.
# This section remains commented out as it requires specific setup (e.g., mingw-w64).
# echo "--- Building Windows Binary (Natively - Requires Setup) ---"
# echo "Building for Windows amd64 (requires mingw-w64 or similar)..."
# CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -v -ldflags="${LDFLAGS_STR}" -o "${OUTPUT_DIR}/goku-windows-amd64.exe" "${MAIN_PACKAGE}"
# echo "----------------------------------------"


# --- Summary ---
echo "Build script finished."
echo "Binaries are located in: ${OUTPUT_DIR}/"
ls -lh "${OUTPUT_DIR}/"