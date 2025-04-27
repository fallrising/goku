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
docker build -t ${IMAGE_NAME} .

echo "Running Docker container to build Linux binary..."
# Run the container:
# --rm: Remove the container after it exits
# -v: Mount the local ./bin directory to /app/bin inside the container
#     The target path (/app/bin) must match where the final Docker stage copies the binary
# ${IMAGE_NAME}: The image to run
# echo "...": A simple command for the container to execute after the build (which happens during image creation)
docker run --rm \
  -v "$(pwd)/${OUTPUT_DIR}:/app/bin" \
  ${IMAGE_NAME} \
  echo "Linux binary build complete (occurred during image build), copied via volume mount."

echo "Linux build process finished."
echo "----------------------------------------"


# --- Build macOS Binaries Natively ---
# These commands run directly on the host machine (your Mac)
# Requires Go and Xcode Command Line Tools to be installed locally
echo "--- Building macOS Binaries (Natively) ---"

echo "Building for macOS amd64..."
# Ensure CGO is enabled
# Add -v for verbose output, helpful for debugging Cgo issues
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -v -ldflags="${LDFLAGS_STR}" -o "${OUTPUT_DIR}/goku-darwin-amd64" "${MAIN_PACKAGE}"
echo "Finished macOS amd64 build."

echo "Building for macOS arm64..."
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -v -ldflags="${LDFLAGS_STR}" -o "${OUTPUT_DIR}/goku-darwin-arm64" "${MAIN_PACKAGE}"
echo "Finished macOS arm64 build."
echo "----------------------------------------"


# --- Build Windows Binary (Placeholder) ---
# CGO cross-compilation from macOS to Windows is non-trivial.
# Requires a specific cross-compiler like mingw-w64 installed on the host.
# This command is commented out as it will likely fail without setup.
# echo "--- Building Windows Binary (Natively - Requires Setup) ---"
# echo "Building for Windows amd64 (requires mingw-w64 or similar)..."
# CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -v -ldflags="${LDFLAGS_STR}" -o "${OUTPUT_DIR}/goku-windows-amd64.exe" "${MAIN_PACKAGE}"
# echo "----------------------------------------"


# --- Summary ---
echo "Build script finished."
echo "Binaries are located in: ${OUTPUT_DIR}/"
ls -lh "${OUTPUT_DIR}/"