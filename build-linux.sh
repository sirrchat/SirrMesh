#!/bin/bash

# Build sirrmeshd for linux-amd64 using Docker

set -e

echo "Building sirrmeshd for linux-amd64..."

# Clean up any existing container
docker rm -f sirrmeshd-extract 2>/dev/null || true

# Build using Docker (with platform emulation on M1)
docker build --platform linux/amd64 -f Dockerfile.build -t sirrmeshd-builder .

# Create a temporary container and copy the binary out
docker create --platform linux/amd64 --name sirrmeshd-extract sirrmeshd-builder
docker cp sirrmeshd-extract:/sirrmeshd ./sirrmeshd-linux-amd64
docker rm sirrmeshd-extract

# Make it executable
chmod +x ./sirrmeshd-linux-amd64

echo ""
echo "Build complete: ./sirrmeshd-linux-amd64"
file ./sirrmeshd-linux-amd64
ls -lh ./sirrmeshd-linux-amd64
