#!/bin/bash

# Create the build directory if it doesn't exist
mkdir -p build

# Build the project
go build -o build/goku-cli ./cmd

# Check if the build was successful
if [ $? -eq 0 ]; then
    echo "Build successful!"
else
    echo "Build failed!"
    exit 1
fi
