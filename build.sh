#!/bin/bash

set -e

echo "Start building Koolo"
echo "Cleaning up previous artifacts..."
rm -rf build || { echo "Error cleaning up previous artifacts"; exit 1; }

echo "Building Koolo binary..."
VERSION=${1:-dev}
GOOS=windows GOARCH=amd64 go build -trimpath -tags static --ldflags "-extldflags=-static -s -w -H windowsgui -X 'github.com/hectorgimenez/koolo/internal/config.Version=$VERSION'" -o build/koolo.exe ./cmd/koolo || { echo "Error building binary"; exit 1; }

echo "Copying assets..."
mkdir -p build/config || { echo "Error creating config directory"; exit 1; }
cp config/koolo.yaml.dist build/config/koolo.yaml || { echo "Error copying koolo.yaml.dist"; exit 1; }
cp config/Settings.json build/config/Settings.json || { echo "Error copying Settings.json"; exit 1; }
cp -r config/template build/config/template || { echo "Error copying template directory"; exit 1; }
cp -r tools build/tools || { echo "Error copying tools directory"; exit 1; }
cp README.md build/ || { echo "Error copying README.md"; exit 1; }

echo "Done! Artifacts are in build directory."
