#!/bin/bash
set -e

echo "Downloading modules..."
go mod download

echo "Generating Prisma Client..."
go run github.com/steebchen/prisma-client-go generate

echo "Building binary..."
go build -o bin/server cmd/api/main.go

echo "Build done."
