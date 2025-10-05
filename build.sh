#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Ensure fyne-cross is installed
echo "Checking for fyne-cross..."
if ! command -v fyne-cross &> /dev/null
then
    echo "fyne-cross could not be found, installing..."
    go install github.com/fyne-io/fyne-cross@latest
fi

echo "Starting the cross-compilation build process..."

# Build for Windows (amd64)
echo "Building for Windows..."
fyne-cross windows -arch amd64 -app-id com.devcode.dosc -name "Devcode ObjectStorage Client" ./cmd/dosc

# Build for macOS (amd64)
echo "Building for macOS..."
fyne-cross darwin -arch amd64 -app-id com.devcode.dosc -name "Devcode ObjectStorage Client" ./cmd/dosc

# Build for Linux (amd64)
echo "Building for Linux (Ubuntu)..."
fyne-cross linux -arch amd64 -app-id com.devcode.dosc -name "Devcode ObjectStorage Client" ./cmd/dosc

echo "Build process completed successfully."
echo "You can find the binaries in the 'fyne-cross/bin' and packages in 'fyne-cross/dist' directories."