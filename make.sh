#!/bin/bash
go build -o bin/wstail *.go
strip bin/wstail
