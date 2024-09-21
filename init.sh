#!/bin/bash

# Copyright (C) 2024 CKC All Rights Reserved.
# 
# File Name: init.sh
# Author   : CKC
# Creation Date: 2024-09-21
# INFO     :
# 
# Create subdirectories
mkdir -p cmd/goku
mkdir -p internal/{bookmarks,tags,search,import,export,database,utils,server,crawl,translate}
mkdir -p pkg/{models,interfaces}
mkdir -p scripts
mkdir -p docs
mkdir -p test/{unit,integration}
mkdir -p assets
mkdir -p config

# Create main.go file
cat << EOF > cmd/goku/main.go
package main

import (
    "fmt"
    "os"
)

func main() {
    fmt.Println("Goku CLI - Bookmark Manager")
    // TODO: Implement CLI logic
}
EOF

# Create go.mod file
go mod init github.com/fallrising/goku-cli

# Create README.md
cat << EOF > README.md
# Goku CLI

Goku CLI is a powerful command-line interface application for managing bookmarks, inspired by the open-source project Buku.

## Features

- Fast and efficient bookmark management
- Advanced search capabilities
- Import and export functionality
- Content analysis and similarity search
- Web crawling
- Translation support

## Installation

TODO: Add installation instructions

## Usage

TODO: Add usage instructions

## Contributing

TODO: Add contribution guidelines

## License

TODO: Add license information
EOF

# Create .gitignore
cat << EOF > .gitignore
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/

# Go workspace file
go.work

# Database files
*.db

# IDE-specific files
.vscode/
.idea/

# OS-specific files
.DS_Store
Thumbs.db
EOF

echo "Goku CLI project structure created successfully!"
