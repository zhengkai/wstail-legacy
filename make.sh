#!/bin/bash

go build -o bin/wstail \
	main.go
strip bin/wstail
