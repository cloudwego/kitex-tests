#! /bin/bash
# Copyright 2021 CloudWeGo Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -e

# add time for each command for timing
export PS4='[$(date "+%F %T")] '

bits=$(getconf LONG_BIT)
if [[ $bits != 64 ]]; then
    echo "this script runs on 64-bit architectures only" >&2
    exit 1
fi

set -x

PATH_BIN=$PWD/bin
mkdir -p $PATH_BIN

export PATH=$PATH_BIN:$PATH

# GOBIN need to be changed here as kitex-tests will install different version bin when testing,
# GOPATH remains the same as default,
# coz other runners in the same host may share path like GOMODCACHE, GOCACHE
export GOBIN=$PATH_BIN

PROTOC_VERSION=v3.20.2

install_protoc() {
    echo "installing protoc ... "
    os=`uname -s | sed 's/Darwin/osx/'`
    arch=`uname -m | sed 's/amd64/x86_64/' | sed 's/arm64/aarch_64/'`
    suffix=$(echo ${os}-${arch})
    filename=protoc-${PROTOC_VERSION#v}-${suffix}.zip
    url=https://github.com/protocolbuffers/protobuf/releases/download/${PROTOC_VERSION}/${filename}
    rm -f $filename
    wget -q $url || exit 1
    unzip -o -q $filename -d ./tmp/
    mv ./tmp/bin/protoc $PATH_BIN
    rm -rf ./tmp/
    echo "installing protoc ... done"
}

go_install() {
    echo "installing $@ ..."
    go install $@ || go get $@
    echo "installing $@ ... done"
}

kitex_cmd() {
  kitex --no-dependency-check $@
}

echo -e "\ninstalling missing commands\n"

# install protoc
which protoc || install_protoc &

# install protoc-gen-go and protoc-gen-go-kitexgrpc
which protoc-gen-go || go_install google.golang.org/protobuf/cmd/protoc-gen-go@latest &

# install protoc-gen-go and protoc-gen-go-kitexgrpc
which protoc-gen-go-grpc || go_install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest &

# install thriftgo
which thriftgo || go_install github.com/cloudwego/thriftgo@latest &

# install kitex
LOCAL_REPO=$1

if [[ -n $LOCAL_REPO ]]; then
    cd ${LOCAL_REPO}
    go_install ${LOCAL_REPO}/tool/cmd/kitex &
    cd -
else
    go_install github.com/cloudwego/kitex/tool/cmd/kitex@develop &
fi

wait
echo -e "\ninstalling missing commands ... done\n"

# double check commands,
# set -e may not working since commands run in background
protoc --version
protoc-gen-go --version
protoc-gen-go-grpc --version
thriftgo --version
kitex -version

rm -f go.mod # go mod init fails if it already exists
go mod init github.com/cloudwego/kitex-tests

echo -e "\ngenerating code for testing ...\n"

rm -rf kitex_gen
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/stability.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/http.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/tenant.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests -combine-service ./idl/combine_service.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/thrift_multi_service.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests ./idl/thrift_multi_service_2.thrift
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/stability.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/unknown_handler.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/grpc_demo.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl -combine-service ./idl/grpc_multi_service.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/grpc_multi_service_2.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl ./idl/pb_multi_service.proto
kitex_cmd -module github.com/cloudwego/kitex-tests -I idl -combine-service ./idl/combine_service.proto

rm -rf kitex_gen_slim
kitex_cmd -module github.com/cloudwego/kitex-tests -thrift template=slim -gen-path kitex_gen_slim ./idl/stability.thrift

rm -rf kitex_gen_noDefSerdes
kitex_cmd -module github.com/cloudwego/kitex-tests -thrift no_default_serdes -gen-path kitex_gen_noDefSerdes ./idl/stability.thrift

rm -rf grpc_gen
mkdir grpc_gen
protoc --go_out=grpc_gen/. ./idl/grpc_demo_2.proto
protoc --go-grpc_out=grpc_gen/. ./idl/grpc_demo_2.proto

# generate thrift streaming code
LOCAL_REPO=$LOCAL_REPO ./thrift_streaming/generate.sh

echo -e "\ngenerating code for testing ... done\n"


echo -e "\nupdating dependencies ... \n"

# Init dependencies
go get github.com/apache/thrift@v0.13.0
go get google.golang.org/grpc@latest
go get google.golang.org/genproto@latest
go get github.com/cloudwego/kitex@develop

if [[ -n $LOCAL_REPO ]]; then
    go mod edit -replace github.com/cloudwego/kitex=${LOCAL_REPO}
fi

go mod edit -replace github.com/cloudwego/kitex=github.com/Marina-Sakai/kitex@v0.9.0-rc7.0.20241010014959-202c253df736
go get github.com/cloudwego/dynamicgo@main

go mod tidy

echo -e "\nupdating dependencies ... done\n"

# run tests
# use less cores for stability
# coz some of them are logical cores
nproc=$(nproc)
export GOMAXPROCS=$(( $nproc / 2 ))

echo -e "\nrunning tests ... \n"

if [[ -n $LOCAL_REPO ]]; then
    go test -failfast -covermode=atomic -coverprofile=${LOCAL_REPO}/coverage.txt -coverpkg=github.com/cloudwego/kitex/... ./...
   if [[ "$OSTYPE" =~ ^darwin ]]; # remove `mode: atomic` line
   then
       sed -i '' 1d ${LOCAL_REPO}/coverage.txt
   else
       sed -i '1d' ${LOCAL_REPO}/coverage.txt
   fi
else
    go test -failfast ./...
fi

echo "PASSED"
