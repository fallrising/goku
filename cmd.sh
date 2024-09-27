#!/bin/bash

# Create the commands directory
mkdir -p cmd/goku/commands

# Create individual command files
touch cmd/goku/commands/add.go
touch cmd/goku/commands/delete.go
touch cmd/goku/commands/get.go
touch cmd/goku/commands/list.go
touch cmd/goku/commands/search.go
touch cmd/goku/commands/update.go
touch cmd/goku/commands/import.go
touch cmd/goku/commands/export.go
touch cmd/goku/commands/tags.go
touch cmd/goku/commands/stats.go

echo "Directory structure and files created successfully."
