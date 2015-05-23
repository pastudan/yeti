#!/bin/bash
PACKAGES=$(find . -maxdepth 1 -mindepth 1 -type d | grep -v .git)
go get -v $PACKAGES
go install -v $PACKAGES
