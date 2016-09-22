#!/bin/bash

# Copyright 2016 Matt Oswalt. Use or modification of this
# source code is governed by the license provided here:
# https://github.com/mierdin/todd/blob/master/LICENSE

# This script configures appropriate capabilities for testlet binaries

set -e
set -u
set -o pipefail

# Ensure GOPATH is set
: ${GOPATH:?"Please ensure GOPATH is set, and run sudo with -E when performing 'make install'"}

# Enable raw socket capabilities on toddping
setcap cap_net_raw+ep $GOPATH/bin/toddping