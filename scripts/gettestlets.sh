#!/bin/bash

# Copyright 2016 Matt Oswalt. Use or modification of this
# source code is governed by the license provided here:
# https://github.com/mierdin/todd/blob/master/LICENSE

# This script installs ToDD-native testlets

set -e
set -u
set -o pipefail

# Install these testlets - comment out specific testlets to
# control what's installed
testlets=(
    'github.com/toddproject/todd-nativetestlet-ping'
)

for i in "${testlets[@]}"
do
   echo "Installing $i"

   # This retrieves from GH if it doesn't exist, but if it does, it installs the local copy.
   go get -u $i/cmd/...
done
