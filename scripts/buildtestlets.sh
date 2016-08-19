#!/bin/bash

# Copyright 2016 Matt Oswalt. Use or modification of this
# source code is governed by the license provided here:
# https://github.com/mierdin/todd/blob/master/LICENSE

# This script downloads ToDD testlets prior to compile

set -e
set -u
set -o pipefail


testlets=(
     'github.com/toddproject/todd-nativetestlet-ping'
   )


#rm -rf testlettemp && mkdir testlettemp && cd testlettemp

for i in "${testlets[@]}"
do
   #echo "Installing $i"
   # git clone $i --quiet
   go get -d -u $i/...
done

# cd ..

# rm -rf agent/testing/downloaded_testlets/ && mkdir agent/testing/downloaded_testlets

# for dir in ./testlettemp/*/
# do
#     dir=${dir%*/}
#     cp testlettemp/${dir##*/}/testlet/* agent/testing/downloaded_testlets
#     #echo ${dir##*/}

    

#     testletdir="$(pwd)/$dir"
#     #echo $testletdir

#     ln -s $testletdir/vendor/ $testletdir/vendor/src

#     # echo ./testlettemp/todd-nativetestlet-ping

#     # Append this vendor directory to GOPATH
#     # TODO need to do some cleanup somewhere to remove this
#     if [[ ":$GOPATH:" != *":$testletdir/vendor:"* ]]; then
#         echo "export GOPATH=$GOPATH:$testletdir/vendor"
#     fi
    

# done




# rm -rf testlettemp



# rebuild plugins:
# _debug "removing: ${plugin_dir:?}/*"
# rm -rf "${plugin_dir:?}/"*
# mkdir -p "${plugin_dir}"

# _info "building plugins"
# find "${__proj_dir}/plugin/" -type d -iname "snap-*" -print0 | xargs -0 -n 1 -I{} "${__dir}/build_plugin.sh" {}

#---------



# __dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# __proj_dir="$(dirname "$__dir")"

# # shellcheck source=scripts/common.sh
# . "${__dir}/common.sh"

# build_dir="${__proj_dir}/build"
# plugin_dir="${build_dir}/plugin"

# plugin_src_path=$1
# plugin_name=$(basename "${plugin_src_path}")
# go_build=(go build -a -ldflags "-w")

# _debug "plugin source: ${plugin_src_path}"
# _info "building ${plugin_name}"

# (cd "${plugin_src_path}" && "${go_build[@]}" -o "${plugin_dir}/${plugin_name}" . || exit 1)