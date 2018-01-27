#!/usr/bin/env bash
echo --------------------------
echo Building the executable
echo --------------------------

CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -a -v -o .build/monday-linux-386
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -a -v -o .build/monday-linux-arm7
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -a -v -o .build/monday-windows-386.exe