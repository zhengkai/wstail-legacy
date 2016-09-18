#!/bin/bash
go build -o bin/wstail *.go
strip bin/wstail
ls -l bin/wstail
cp bin/wstail release/
