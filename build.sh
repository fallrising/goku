#!/bin/bash

set -e

# Ensure bin directory exists
mkdir -p bin

echo "Building Goku CLI..."
go build -o bin/goku cmd/goku/main.go

echo "Build completed successfully. Binary is located at bin/goku"