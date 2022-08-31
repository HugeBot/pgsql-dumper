#!/bin/sh

echo "Compiling for Windows amd64..."
GOOS=windows GOARCH=amd64 go build -o bin/psql-dumper-x86_64.exe -ldflags "-X main.Version=$RELEASE_TAG" main.go

echo "Compiling for Windows i386..."
GOOS=windows GOARCH=386 go build -o bin/psql-dumper-i386.exe -ldflags "-X main.Version=$RELEASE_TAG" main.go

echo "Compiling for Darwin amd64..."
GOOS=darwin GOARCH=amd64 go build -o bin/psql-dumper-x86_64-darwin -ldflags "-X main.Version=$RELEASE_TAG" main.go

echo "Compiling for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -o bin/psql-dumper-x86_64-linux -ldflags "-X main.Version=$RELEASE_TAG" main.go

echo "Compiling for Linux i386..."
GOOS=linux GOARCH=386 go build -o bin/psql-dumper-i386-linux -ldflags "-X main.Version=$RELEASE_TAG" main.go

echo "Done!!"