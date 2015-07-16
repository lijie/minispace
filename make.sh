#!/bin/sh

export GOPATH=/home/lijie/projects/gocode
go install -x -tags "usemgo" github.com/lijie/minispace/server/minispaced
