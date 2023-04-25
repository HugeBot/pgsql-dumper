#!/bin/sh

echo "Compiling for Windows amd64..."
GOOS=windows GOARCH=amd64 go build -o bin/pgsql-dumper-amd64.exe -ldflags "-X main.Version=$RELEASE_TAG" main.go 

echo "Compiling for Windows i386..."
GOOS=windows GOARCH=386 go build -o bin/pgsql-dumper-i386.exe -ldflags "-X main.Version=$RELEASE_TAG" main.go

echo "Compiling for Darwin amd64..."
GOOS=darwin GOARCH=amd64 go build -o bin/pgsql-dumper-amd64-darwin -ldflags "-X main.Version=$RELEASE_TAG" main.go
chmod +x bin/pgsql-dumper-amd64-darwin

echo "Compiling for Linux amd64..."
GOOS=linux GOARCH=amd64 go build -o bin/pgsql-dumper-amd64-linux -ldflags "-X main.Version=$RELEASE_TAG" main.go
chmod +x bin/pgsql-dumper-amd64-linux

echo "Compiling for Linux i386..."
GOOS=linux GOARCH=386 go build -o bin/pgsql-dumper-i386-linux -ldflags "-X main.Version=$RELEASE_TAG" main.go
chmod +x bin/pgsql-dumper-i386-linux

echo "Done!!"