#!/bin/bash
set -e

echo "Building Laravel Deploy Panel..."

# Build frontend
cd frontend
npm install
npm run build
cd ..

# Build Go binary
go build -ldflags="-s -w" -o panel .

echo "Build complete: ./panel"
echo "Run with: sudo ./panel [--port 4432]"
